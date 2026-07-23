package proto

import (
	"encoding/json"
	"larsys/go/lib"
	"net"
)

func writeFramed(conn net.Conn, v any, logger *lib.Logger) error {
	if logger != nil {
		logger.Println(v)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(b, '\n'))
	return err
}

type Node struct {
	IP       string
	PORT     int
	LOG      lib.LoggingConfig
	USERNAME string
	DEVICE   string
}
