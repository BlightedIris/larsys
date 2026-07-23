package registry

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"larsys/go/lib"
	"larsys/go/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var RELEASES_DIR = filepath.Join(proto.PLUGIN_DIR, "registry", "releases")

func main() {
	var reg proto.Node
	flag.StringVar(&reg.IP, "ip", "localhost", "IP address")
	flag.IntVar(&reg.PORT, "port", 5450, "Port")
	flag.StringVar(&reg.LOG.PATH, "log", "/var/log/larsys/plugins/registry/daemon.log", "Path to daemon log file")
	flag.StringVar(&reg.LOG.LEVEL, "level", "debug", "LogLevel")
	flag.StringVar(&reg.USERNAME, "username", proto.GetUsername(), "OS username")
	flag.Parse()

	reg.DEVICE = "REGISTRY"

	// --- Folders
	proto.InitDirs()
	os.MkdirAll(RELEASES_DIR, 0o755)
	logger := lib.GetLogger(reg.LOG.PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, log.Ldate|log.Ltime)
	defer logger.Close()
	addr := fmt.Sprintf("%s:%d", reg.IP, reg.PORT)
	responder := proto.Responder{LOGGER: logger, SRC: addr}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	defer listener.Close()
	logger.Printf("Listening on %s", addr)
	logger.Printf("Running as %s on %s", reg.USERNAME, reg.DEVICE)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

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

	sig := <-stop
	logger.Printf("Received signal: %v", sig)
	listener.Close()
	logger.Println("Shutting down...")
}
func checkRequest(req proto.Request) (error, string, string) {
	if params, ok := req.ACTION.PARAMS.(map[string]string); ok {
		if name, exists := params["NAME"]; exists {
			if version, exists := params["VERSION"]; exists {
				return nil, name, version
			} else {
				return errors.New("Plugin version was not found"), "", ""
			}
		} else {
			return errors.New("Plugin was not found"), "", ""
		}
	} else {
		return errors.New("PARAMS is wrong type or nil"), "", ""
	}
}

func handleConnection(conn net.Conn, responder proto.Responder, logger *lib.Logger) {
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
		if proto.ActionMatches(req.ACTION, proto.PING) {
			resp := proto.Response{SRC: responder.SRC, STATUS: 0, MSG: "Pong"}
			responder.Respond(conn, resp)
			break
		} else if req.ACTION.NAME == "plugin/download" {
			if err, name, version := checkRequest(req); err == nil {
				filePath := filepath.Join(RELEASES_DIR, name, version)
				data, err := os.ReadFile(filePath)
				if err != nil {
					responder.Error(conn, "Something went wrong opening the file", err)
					break
				}
				responder.Respond(conn, proto.Response{
					STATUS: 0,
					MSG:    "OK",
					PARAMS: proto.PLUGIN_DOWNLOAD_PARAMS{
						NAME: filepath.Base(filePath), // just the filename
						FILE: data,
					},
				})
			} else {
				responder.Error(conn, "", err)
			}
		} else if req.ACTION.NAME == "plugin/info" {
			if err, name, version := checkRequest(req); err == nil {
				filePath := filepath.Join(RELEASES_DIR, name, version, "manifest.toml")
				data, err := os.ReadFile(filePath)
				if err != nil {
					responder.Error(conn, "Something went wrong opening the file", err)
					break
				}
				responder.Respond(conn, proto.Response{
					STATUS: 0,
					MSG:    "OK",
					PARAMS: proto.PLUGIN_DOWNLOAD_PARAMS{
						NAME: filepath.Base(filePath), // just the filename
						FILE: data,
					},
				})
			} else {
				responder.Error(conn, "", err)
			}
		} else if req.ACTION.NAME == "plugin/upload" {
			if proto.IsRegistered(req) {
				if err, name, version := checkRequest(req); err == nil {
					path := filepath.Join(RELEASES_DIR, name, version)
					if info, err := os.Stat(path); err == nil && info.IsDir() {
						responder.Error(conn, "Plugin version already exists", err)
						break
					}
					err := os.MkdirAll(path, 0o755)
					if err != nil {
						responder.Error(conn, "", err)
					}

					zipPath := filepath.Join(path, name+".zip") // write to the zip file path

					if params, ok := req.ACTION.PARAMS.(map[string]any); ok {
						if data, ok := params["FILE"].([]byte); ok {
							os.WriteFile(zipPath, data, 0644) // write to zipPath
						}
					}

					reader, err := zip.OpenReader(zipPath) // open the zip file, not the directory
					if err != nil {
						responder.Error(conn, "", err)
						break
					}
					defer reader.Close()

					for _, file := range reader.File {
						outPath := filepath.Join(path, file.Name)
						rc, err := file.Open()
						if err != nil {
							responder.Error(conn, "Failed to open file in archive:", err)
							break
						}
						outFile, _ := os.Create(outPath)
						io.Copy(outFile, rc)
						outFile.Close()
						rc.Close()
					}
					break
				} else {
					responder.Error(conn, "", err)
				}
			} else {
				responder.Unauthorised(conn)
			}
		} else {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("Scanner error: %v", err)
	}
	logger.Printf("Client disconnected: %s", addr)
}
