package main

import (
	"crypto/sha256"
	"encoding/hex"
	"larsys/go/proto"
	"os"
	"path/filepath"
)

func register(req proto.Request) (bool, error) {
	tokenDir := filepath.Join(proto.TOKEN_DIR, "clients", req.SRC)
	os.MkdirAll(tokenDir, 0o755)
	os.Chown(tokenDir, os.Geteuid(), os.Getegid())
	tokenFile := proto.GetTokenPath(req.SRC, false)
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

func is_registered(req proto.Request) bool {
	_, err := os.Stat(proto.GetTokenPath(req.SRC, false))
	return err == nil
}

func authorise(req proto.Request) bool {
	tokenBytes, err := os.ReadFile(proto.GetTokenPath(req.SRC, false))
	if err != nil {
		return false
	}
	return req.TOKEN == string(tokenBytes)
}

func revoke(req proto.Request) error {
	tokenPath := proto.GetTokenPath(req.SRC, false)
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
