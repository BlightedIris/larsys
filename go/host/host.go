package main

import (
	"bufio"
	"flag"
	"fmt"
	"larsys/go/shared"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

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
