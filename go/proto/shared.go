package proto

import (
	"os"
	"os/user"
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

func GetUsername() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u.Username
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

func ActionMatches(actual Action, expected Action) bool {
	return actual.NAME == expected.NAME
}
