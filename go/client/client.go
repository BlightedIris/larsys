package main

import (
	"bufio"
	"flag"
	"fmt"
	"larsys/go/shared"
	"net"
	"os"
)

func main() {
	// --- Flags
	var client_conf shared.Node
	flag.StringVar(&client_conf.IP, "client-ip", "localhost", "Client IP address")
	flag.IntVar(&client_conf.PORT, "client-port", 5453, "Client port")
	flag.StringVar(&client_conf.LOG.PATH, "log", "/var/log/larsys/daemon.log", "Path to daemon log file")
	flag.StringVar(&client_conf.LOG.LEVEL, "level", "debug", "LogLevel")

	// --- Hard Values
	client_conf.LOG.NAME = "CLIENT DAEMON"
	client_conf.LOG.RULES = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// --- Host ref
	host_ip := flag.String("host-ip", "localhost", "Host IP")
	host_port := flag.Int("host-port", 5454, "Host port")
	host := fmt.Sprintf("%s:%d", *host_ip, *host_port)

	flag.Parse()

	// --- Message reader
	for {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			input := scanner.Text()
			go send_message(input, host)
		}

		if scanner.Err() != nil {
			panic("Something went wrong")
		}
	}
}

func send_message(input string, host string) {
	// Connect to the host
	conn, err := net.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Send message
	fmt.Fprintf(conn, "%s\n", input)

	// Read response
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		response := scanner.Text()
		fmt.Println("Response:", response)
	}
}
