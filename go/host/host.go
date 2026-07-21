package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"larsys/go/shared"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// This file should be called by the larsys user.
var LARSYS_USER_UID = os.Getuid()
var LARSYS_USER_GID = os.Getgid()
var LARSYS_SRC string

func main() {
	//  --- Flags
	var host_conf shared.Node
	flag.StringVar(&host_conf.IP, "host", "localhost", "IP address of the host")
	flag.IntVar(&host_conf.PORT, "port", 5454, "Port the host will be listening to")
	flag.StringVar(&host_conf.LOG.PATH, "log", "/var/log/larsys/daemon.log", "Path to daemon log file")
	flag.StringVar(&host_conf.LOG.LEVEL, "level", "debug", "LogLevel")
	flag.StringVar(&host_conf.USERNAME, "username", shared.GetUsername(), "OS username")
	flag.StringVar(&host_conf.DEVICE, "device", "ADMIN MACHINE", "Name of the device running the service")
	flag.Parse()

	// --- Hard Values
	host_conf.LOG.NAME = "HOST DAEMON"
	host_conf.LOG.RULES = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// --- Folders
	shared.InitDirs()

	// --- Logging
	log_f, err := os.OpenFile(host_conf.LOG.PATH, host_conf.LOG.RULES, 0o644)
	if err != nil {
		panic(err)
	}
	defer log_f.Close()
	logger := log.New(log_f, "", log.Ldate|log.Ltime)
	logger.Println("Starting Daemon...")

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
			go handleConnection(conn, logger)
		}
	}()

	// --- Exit
	sig := <-stop
	logger.Printf("Received signal: %v", sig)
	listener.Close()
	logger.Println("Shutting down...")
}

func is_registered(req shared.Request) bool {
	_, err := os.Stat(shared.GetTokenPath(req.SRC, false))
	return err == nil
}

func authorise(req shared.Request) bool {
	tokenBytes, err := os.ReadFile(shared.GetTokenPath(req.SRC, false))
	if err != nil {
		return false
	}
	return req.TOKEN == string(tokenBytes)
}

func actionMatches(actual shared.Action, expected shared.Action) bool {
	return actual.NAME == expected.NAME
}
func register(req shared.Request) (bool, error) {
	tokenDir := filepath.Join(shared.TOKEN_DIR, "clients", req.SRC)
	os.MkdirAll(tokenDir, 0o755)
	os.Chown(tokenDir, os.Geteuid(), os.Getegid())
	tokenFile := shared.GetTokenPath(req.SRC, false)
	tokenHash := sha256.Sum256([]byte(tokenFile))
	hashString := hex.EncodeToString(tokenHash[:])
	err := os.WriteFile(tokenFile, []byte(hashString), 0400)
	if err != nil {
		return false, err
	}
	permErr := os.Chown(tokenFile, os.Geteuid(), os.Getegid())
	if permErr != nil {
		return false, permErr
	}
	return true, nil
}
func handleConnection(conn net.Conn, logger *log.Logger) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	logger.Printf("Client connected: %s", addr)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req shared.Request
		err := json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			logger.Printf("%s", err)
			resp := shared.Response{SRC: LARSYS_SRC, STATUS: 1, MSG: "An error ocurred"}
			respBytes, _ := json.Marshal(resp)
			conn.Write(respBytes)
			break
		}
		if actionMatches(req.ACTION, shared.PING) {
			resp := shared.Response{SRC: LARSYS_SRC, STATUS: 0, MSG: "Pong"}
			respBytes, _ := json.Marshal(resp)
			conn.Write(respBytes)
			break
		} else if actionMatches(req.ACTION, shared.REGISTER) {
			if !is_registered(req) {
				ok, err := register(req)
				if err != nil || !ok {
					logger.Printf("%s", err)
					resp := shared.Response{SRC: LARSYS_SRC, STATUS: 1, MSG: "An error ocurred while registering the client"}
					respBytes, _ := json.Marshal(resp)
					conn.Write(respBytes)
				} else {
					token, _ := os.ReadFile(shared.GetTokenPath(req.SRC, false))
					logger.Printf("+++ Client Registered: %s +++", req.SRC)
					resp := shared.REGISTERED
					resp.SRC = LARSYS_SRC
					resp.PARAMS = shared.REGISTERED_PARAMS{TOKEN: string(token)}
					respBytes, _ := json.Marshal(resp)
					conn.Write(respBytes)
				}
			} else {
				logger.Printf("Client %s is already registered", req.SRC)
				resp := shared.Response{SRC: LARSYS_SRC, STATUS: 1, MSG: "Client already registered."}
				respBytes, _ := json.Marshal(resp)
				conn.Write(respBytes)
			}
			break
		} else if authorise(req) {
			logger.Printf("Client %s has been recognised.", req.SRC)
			logger.Printf("### Authentication succesfull: %s ###", req.SRC)

			switch req.ACTION.NAME {
			case shared.REVOKE.NAME:
				tokenPath := shared.GetTokenPath(req.SRC, false)
				_, err := os.Stat(tokenPath)
				if err != nil {
					resp := shared.Response{SRC: LARSYS_SRC, STATUS: 1, MSG: "Token not found"}
					respBytes, _ := json.Marshal(resp)
					conn.Write(respBytes)
					break
				}
				delErr := os.Remove(tokenPath)
				if delErr != nil {
					resp := shared.Response{SRC: LARSYS_SRC, STATUS: 1, MSG: "Failed to delete token"}
					respBytes, _ := json.Marshal(resp)
					conn.Write(respBytes)
					break
				}
				logger.Printf("--- Deleted Client: %s ---", req.SRC)
				// case shared.PLUGIN_INSTALL:
				// 	install_plugin(req)
				// case shared.PLUGIN_UNINSTALL:
				// 	uninstall_plugin(req)
			}
			respBytes, _ := json.Marshal(shared.OK)
			conn.Write(respBytes)
			break
		} else {
			logger.Println("Unauthorised access attempt:")
			logger.Printf("SRC: %s", req.SRC)
			logger.Printf("TOKEN: %s", req.TOKEN)
			logger.Printf("ACTION: %s", req.ACTION.NAME)

			respBytes, _ := json.Marshal(shared.UNAUTHORISED)
			conn.Write(respBytes)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("Scanner error: %v", err)
	}

	logger.Printf("Client disconnected: %s", addr)
}
