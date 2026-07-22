package proto

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
)

var PLUGIN_DIR = "/opt/larsys/plugins"
var TOKEN_DIR = "/etc/larsys/tokens"
var TOKEN_DIR_CL = "/etc/larsys/tokens/clients"
var TOKEN_DIR_HS = "/etc/larsys/tokens/hosts"
var LOG_DIR = "/var/log/larsys/"
var STATE_DIR = "/var/lib/larsys/"

func InitDirs() {
	dirs := []string{
		PLUGIN_DIR,
		TOKEN_DIR_CL,
		TOKEN_DIR_HS,
		LOG_DIR,
		STATE_DIR,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			panic(err)
		}
	}
}

type LoggingConfig struct {
	PATH  string
	RULES int
	LEVEL string
	NAME  string
}

type Node struct {
	IP       string
	PORT     int
	LOG      LoggingConfig
	USERNAME string
	DEVICE   string
}

func GetUsername() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u.Username
}

func GetTokenPath(name string, host bool) string {
	var dest string
	if host {
		dest = "hosts"
	} else {
		dest = "clients"
	}
	return filepath.Join(TOKEN_DIR, dest, name, "token")
}

func ExtractParam(obj map[string]any, param string) any {
	if obj == nil {
		return ""
	}

	if value, exists := obj[param]; exists {
		return value
	}

	return nil
}

func SaveToken(host, token string) error {
	tokenDir := filepath.Join(TOKEN_DIR, "hosts", host)
	if err := os.MkdirAll(tokenDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(tokenDir, "token"), []byte(token), 0o600)
}

func writeFramed(conn net.Conn, v any, logger *log.Logger) error {
	logger.Println(v)
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(b, '\n'))
	return err
}
