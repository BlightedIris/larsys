package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type NetworkConfig struct {
	host string
	port int
}

type LoggingConfig struct {
	file  string
	level string
	name  string
}

type AppConfig struct {
	max_clients int
}

func main() {
	var network_conf NetworkConfig
	flag.StringVar(&network_conf.host, "host", "localhost", "IP address of the host")
	flag.IntVar(&network_conf.port, "port", 5454, "Port the host will be listening to")

	var logging_conf LoggingConfig
	flag.StringVar(&logging_conf.file, "log", "/var/log/larsys/daemon.log", "Path to daemon log file")
	flag.StringVar(&logging_conf.level, "level", "debug", "LogLevel")
	logging_conf.name = "DAEMON"

	var app_conf AppConfig
	flag.IntVar(&app_conf.max_clients, "max-clients", 2, "Max number of clients")

	flag.Parse()

	log_f, err := os.Create(logging_conf.file)
	if err != nil {
		panic(err)
	}
	defer log_f.Close()

	logger := log.New(log_f, "", log.Ldate|log.Ltime)
	logger.Println("Starting Daemon...")

	addr := fmt.Sprintf("%s:%d", network_conf.host, network_conf.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	defer listener.Close()

	logger.Printf("Listening on %s", addr)

	// Channel to signal shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

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

	sig := <-stop // blocks here, waiting for Ctrl+C
	logger.Printf("Received signal: %v", sig)
	listener.Close()
	logger.Println("Shutting down...")
}

// func ui_shutdown() {

// }

func handleConnection(conn net.Conn, logger *log.Logger) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	logger.Printf("Client connected: %s", addr)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		logger.Printf("Received: %s", message)
		fmt.Fprintf(conn, "Echo: %s\n", message) // send back to client
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("Scanner error: %v", err)
	}

	logger.Printf("Client disconnected: %s", addr)
}
