package proto

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

func GetTokenPath(name string, host bool) string {
	var dest string
	if host {
		dest = "hosts"
	} else {
		dest = "clients"
	}
	return filepath.Join(TOKEN_DIR, dest, name, "token")
}

func SaveToken(host, token string) error {
	tokenDir := filepath.Join(TOKEN_DIR, "hosts", host)
	if err := os.MkdirAll(tokenDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(tokenDir, "token"), []byte(token), 0o600)
}

func IsRegistered(req Request) bool {
	_, err := os.Stat(GetTokenPath(req.SRC, false))
	return err == nil
}
func Register(req Request) (bool, error) {
	tokenDir := filepath.Join(TOKEN_DIR, "clients", req.SRC)
	os.MkdirAll(tokenDir, 0o755)
	os.Chown(tokenDir, os.Geteuid(), os.Getegid())
	tokenFile := GetTokenPath(req.SRC, false)
	tokenHash := sha256.Sum256([]byte(tokenFile))
	hashString := hex.EncodeToString(tokenHash[:])
	err := os.WriteFile(tokenFile, []byte(hashString), 0400)
	if err != nil {
		return false, err
	}
	// Only attempt to change ownership if running as root
	if os.Geteuid() == 0 {
		permErr := os.Chown(tokenFile, os.Geteuid(), os.Getegid())
		if permErr != nil {
			return false, permErr
		}
	}
	return true, nil
}

func Authorise(req Request) bool {
	tokenBytes, err := os.ReadFile(GetTokenPath(req.SRC, false))
	if err != nil {
		return false
	}
	return req.TOKEN == string(tokenBytes)
}

func Revoke(req Request) error {
	tokenPath := GetTokenPath(req.SRC, false)
	_, err := os.Stat(tokenPath)
	if err != nil {
		return err
	}
	delErr := os.Remove(tokenPath)
	if delErr != nil {
		return err
	}
	return nil
}
