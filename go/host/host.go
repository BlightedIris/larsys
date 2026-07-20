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

var PLUGIN_DIR = "/opt/larsys/plugins"
var TOKEN_DIR = "/etc/larsys/tokens"
var LOG_DIR = "/var/log/larsys/"
var STATE_DIR = "/var/lib/larsys/"

func init_dirs() {
	dirs := []string{
		PLUGIN_DIR,
		TOKEN_DIR,
		LOG_DIR,
		STATE_DIR,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			panic(err)
		}
	}
}

func main() {
	//  --- Flags
	var host_conf shared.Node
	flag.StringVar(&host_conf.IP, "host", "localhost", "IP address of the host")
	flag.IntVar(&host_conf.PORT, "port", 5454, "Port the host will be listening to")
	flag.StringVar(&host_conf.LOG.PATH, "log", "/var/log/larsys/daemon.log", "Path to daemon log file")
	flag.StringVar(&host_conf.LOG.LEVEL, "level", "debug", "LogLevel")
	flag.Parse()

	// --- Hard Values
	host_conf.LOG.NAME = "HOST DAEMON"
	host_conf.LOG.RULES = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// --- Folders
	init_dirs()

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
	clientDir := filepath.Join(TOKEN_DIR, req.SRC)
	info, err := os.Stat(clientDir)
	if err != nil {
		return false
	}

	tokenBytes, err := os.ReadFile(filepath.Join(clientDir, "token"))
	if err != nil {
		return false
	}
	if info.IsDir() && req.SRC == string(tokenBytes) {
		return true
	} else {
		return false
	}
}

func register(req shared.Request) bool {
	token := filepath.Join(TOKEN_DIR, req.SRC, "token")
	tokenHash := sha256.Sum256([]byte(token))
	hashString := hex.EncodeToString(tokenHash[:])
	err := os.WriteFile(token, []byte(hashString), 0400)
	if err != nil{
		return false
	}
	return  true
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
			resp := shared.Response{STATUS: 1, MSG: "An error ocurred"}
			respBytes, _ := json.Marshal(resp)
			conn.Write(respBytes)
			return
		}
		if is_registered(req) {
			if req.ACTION == "register" {
				respBytes, err := json.Marshal(shared.ALREADY_REGISTERED)
				if err != nil {
					resp := shared.Response{STATUS: 1, MSG: "An error ocurred"}
					respBytes, _ := json.Marshal(resp)
					conn.Write(respBytes)
					return
				}
				conn.Write(respBytes)
				return
			}

			// TODO: execute action

		} else if req.ACTION == "register" {
			if !register(req) {
				resp := shared.Response{STATUS: 1, MSG: "An error ocurred while registering the client"}
				respBytes, _ := json.Marshal(resp)
				conn.Write(respBytes)
				return
			}
		} else {
			respBytes, _ := json.Marshal(shared.UNAUTHORISED)
			conn.Write(respBytes)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("Scanner error: %v", err)
	}

	logger.Printf("Client disconnected: %s", addr)
}
