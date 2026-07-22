package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"larsys/go/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// This file should be called by the larsys user.
var LARSYS_USER_UID = os.Getuid()
var LARSYS_USER_GID = os.Getgid()
var LARSYS_SRC string

func main() {
	//  --- Flags
	var host_conf proto.Node
	flag.StringVar(&host_conf.IP, "host", "localhost", "IP address of the host")
	flag.IntVar(&host_conf.PORT, "port", 5454, "Port the host will be listening to")
	flag.StringVar(&host_conf.LOG.PATH, "log", "/var/log/larsys/daemon.log", "Path to daemon log file")
	flag.StringVar(&host_conf.LOG.LEVEL, "level", "debug", "LogLevel")
	flag.StringVar(&host_conf.USERNAME, "username", proto.GetUsername(), "OS username")
	flag.StringVar(&host_conf.DEVICE, "device", "ADMIN MACHINE", "Name of the device running the service")
	flag.Parse()

	// --- Hard Values
	host_conf.LOG.NAME = "HOST DAEMON"
	host_conf.LOG.RULES = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// --- Folders
	proto.InitDirs()

	// --- Logging
	log_f, err := os.OpenFile(host_conf.LOG.PATH, host_conf.LOG.RULES, 0o644)
	if err != nil {
		panic(err)
	}
	defer log_f.Close()
	logger := log.New(log_f, "", log.Ldate|log.Ltime)
	logger.Println("Starting Daemon...")

	responder := proto.Responder{
		LOGGER: logger,
		SRC:    LARSYS_SRC,
	}

	// --- Listener
	addr := fmt.Sprintf("%s:%d", host_conf.IP, host_conf.PORT)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	defer listener.Close()
	logger.Printf("Listening on %s", addr)
	logger.Printf("Running as %s on %s", host_conf.USERNAME, host_conf.DEVICE)
	LARSYS_SRC = fmt.Sprintf("%s:%s", host_conf.USERNAME, host_conf.DEVICE)
	// --- Channels
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// --- Messaging
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Printf("Accept error: %v", err)
				return // exit the loop
			}
			go handleConnection(conn, responder, logger)
		}
	}()

	// --- Exit
	sig := <-stop
	logger.Printf("Received signal: %v", sig)
	listener.Close()
	logger.Println("Shutting down...")
}

func actionMatches(actual proto.Action, expected proto.Action) bool {
	return actual.NAME == expected.NAME
}

func handleConnection(conn net.Conn, responder proto.Responder, logger *log.Logger) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	logger.Printf("Client connected: %s", addr)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req proto.Request
		err := json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			responder.Error(conn, "", err)
			break
		}
		if actionMatches(req.ACTION, proto.PING) {
			resp := proto.Response{SRC: LARSYS_SRC, STATUS: 0, MSG: "Pong"}
			responder.Respond(conn, resp)
			break
		} else if actionMatches(req.ACTION, proto.REGISTER) {
			if !is_registered(req) {
				ok, err := register(req)
				if err != nil || !ok {
					responder.Error(conn, "An error ocurred when registering the client", err)
				} else {
					token, _ := os.ReadFile(proto.GetTokenPath(req.SRC, false))
					responder.Register(conn, string(token))
				}
			} else {
				responder.AlreadyRegistered(conn)
			}
			break
		} else if authorise(req) {
			logger.Printf("Client %s has been recognised.", req.SRC)
			logger.Printf("### Authentication succesfull: %s ###", req.SRC)
			switch req.ACTION.NAME {
			case proto.REVOKE.NAME:
				err := revoke(req)
				if err != nil {
					responder.Error(conn, "An error ocurred while deleting the token", err)
				}
				responder.OK(conn, fmt.Sprintf("--- Deleted Client: %s ---", req.SRC))
				// case proto.PLUGIN_INSTALL:
				// 	install_plugin(req)
				// case proto.PLUGIN_UNINSTALL:
				// 	uninstall_plugin(req)
			}
			break
		} else {
			logger.Println("Unauthorised access attempt:")
			logger.Printf("SRC: %s", req.SRC)
			logger.Printf("TOKEN: %s", req.TOKEN)
			logger.Printf("ACTION: %s", req.ACTION.NAME)
			responder.Unauthorised(conn)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("Scanner error: %v", err)
	}

	logger.Printf("Client disconnected: %s", addr)
}
