package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"larsys/go/lib"
	"larsys/go/proto"
	"log"
	"os"
)

func main() {
	// --- Flags
	var client_conf proto.Node
	flag.StringVar(&client_conf.IP, "client-ip", "localhost", "Client IP address")
	flag.IntVar(&client_conf.PORT, "client-port", 5453, "Client port")
	flag.StringVar(&client_conf.LOG.PATH, "log", "/var/log/larsys/daemon.log", "Path to daemon log file")
	flag.StringVar(&client_conf.LOG.LEVEL, "level", "debug", "LogLevel")
	flag.StringVar(&client_conf.USERNAME, "username", proto.GetUsername(), "OS username")
	flag.StringVar(&client_conf.DEVICE, "device", "CLIENT MACHINE", "Name of the device running the service")

	// --- Hard coded Values
	client_conf.LOG.NAME = "CLIENT DAEMON"
	client_conf.LOG.RULES = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// --- Host ref
	host_ip := flag.String("host-ip", "localhost", "Host IP")
	host_port := flag.Int("host-port", 5454, "Host port")
	host_name := flag.String("host-name", "root:ADMIN MACHINE", "Name of the host")

	var action proto.Action
	// --- CLI actions
	flag.StringVar(&action.NAME, "action", "ping", "Action to execute")
	var paramsStr string
	flag.StringVar(&paramsStr, "params", `{}`, "Json in string")
	err := json.Unmarshal([]byte(paramsStr), &action.PARAMS)
	if err != nil {
		panic(err)
	}
	flag.Parse()

	// --- Folders
	proto.InitDirs()

	// --- Logging
	logger := lib.GetLogger(client_conf.LOG.PATH, client_conf.LOG.RULES, log.Ldate|log.Ltime)
	defer logger.Close()

	logger.Printf("Hosting at %s: %s:%d\n", *host_name, *host_ip, *host_port)
	dest := make(map[string]proto.Host)
	dest[*host_name] = proto.Host{
		IP:   *host_ip,
		PORT: *host_port,
	}
	sender := proto.Sender{
		LOGGER: logger,
		SRC:    fmt.Sprintf("%s:%s", client_conf.USERNAME, client_conf.DEVICE),
		TOKENS: make(map[string]string),
		DST:    dest}

	// --- Message reader
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var resp proto.Response
		var err error
		input := scanner.Text()
		switch input {
		case "ping":
			resp, err = sender.Ping(*host_name)
		case "register":
			resp, err = sender.Register(*host_name)
		case "revoke":
			resp, err = sender.Revoke(*host_name)
		case "plugin/install":
			resp, err = sender.PluginInstall(*host_name, proto.PLUGIN{})
		case "plugin/uninstall":
			resp, err = sender.PluginUninstall(*host_name, proto.PLUGIN{})
		default:
			continue
		}
		if err != nil {
			logger.Println(err)
		}
		logger.Println(resp)
	}

	if scanner.Err() != nil {
		panic("Something went wrong")
	}
}
