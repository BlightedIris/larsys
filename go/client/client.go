package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"larsys/go/proto"
	"net"
	"os"
)

var CLIENT_SRC string

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

	host := fmt.Sprintf("%s:%d", *host_ip, *host_port)
	fmt.Printf("Hosting at %s: %s:%d\n", *host_name, *host_ip, *host_port)
	CLIENT_SRC = fmt.Sprintf("%s:%s", client_conf.USERNAME, client_conf.DEVICE)
	// --- Message reader
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		var req proto.Request
		switch input {
		case "ping":
			req = ping_request(CLIENT_SRC, client_conf.DEVICE)
		case "register":
			req = register_request(CLIENT_SRC, action.PARAMS)
		case "revoke":
			req = revoke_request(CLIENT_SRC, action.PARAMS)
			tokenBytes, err := os.ReadFile(proto.GetTokenPath(*host_name, true))

			if err != nil {
				fmt.Printf("Something went wrong: %s", err)
			}

			req.TOKEN = string(tokenBytes)

		case "plugin/install":
			req = plugin_install_request(CLIENT_SRC, action.PARAMS)
		case "plugin/uninstall":
			req = plugin_uninstall_request(CLIENT_SRC, action.PARAMS)
		default:
			continue
		}

		send_message(req, host, *host_name)
	}

	if scanner.Err() != nil {
		panic("Something went wrong")
	}
}

func ping_request(src, device string) proto.Request {
	return proto.Request{
		SRC: src,
		ACTION: proto.Action{
			NAME:   "ping",
			PARAMS: map[string]string{"DEVICE": device},
		},
	}
}

func register_request(src string, params any) proto.Request {
	return proto.Request{
		SRC: src,
		ACTION: proto.Action{
			NAME:   "register",
			PARAMS: params,
		},
	}
}

func revoke_request(src string, params any) proto.Request {
	return proto.Request{
		SRC: src,
		ACTION: proto.Action{
			NAME:   "revoke",
			PARAMS: params,
		},
	}
}

func plugin_install_request(src string, params any) proto.Request {
	return proto.Request{
		SRC: src,
		ACTION: proto.Action{
			NAME:   "plugin/install",
			PARAMS: params,
		},
	}
}

func plugin_uninstall_request(src string, params any) proto.Request {
	return proto.Request{
		SRC: src,
		ACTION: proto.Action{
			NAME:   "plugin/uninstall",
			PARAMS: params,
		},
	}
}

func send_message(req proto.Request, host string, host_name string) *proto.Response {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	token_path := proto.GetTokenPath(host_name, true)
	tkn, err := os.ReadFile(token_path)
	req.TOKEN = string(tkn)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	if _, err := conn.Write(append(reqBytes, '\n')); err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		response := scanner.Bytes()
		fmt.Println("Response:", string(response))

		var resp proto.Response
		if err := json.Unmarshal(response, &resp); err != nil {
			fmt.Println("response parse failed:", err)
			return nil
		}

		if params, ok := resp.PARAMS.(map[string]any); ok {
			if token, ok := params["TOKEN"].(string); ok {
				if err := proto.SaveToken(host_name, token); err != nil {
					fmt.Println("save token failed:", err)
				} else {
					fmt.Println("Token saved for", host_name)
				}
			}
		}

		return &resp
	} else if err := scanner.Err(); err != nil {
		fmt.Println("read failed:", err)
	} else {
		fmt.Println("no response received")
	}

	return nil
}
