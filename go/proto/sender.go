package proto

import (
	"errors"
	"log"
	"net"
	"os"
)

type Request struct {
	SRC    string
	DST    string
	TOKEN  string
	ACTION Action
}

type Sender struct {
	LOGGER *log.Logger
	SRC    string
	TOKENS map[string]string
	DST    map[string]string
}

func NewSender(src string, tokenPath map[string]string, logger *log.Logger, destination map[string]string) *Sender {
	return &Sender{
		DST:    destination,
		LOGGER: logger,
		SRC:    src,
		TOKENS: tokenPath,
	}
}

func (s *Sender) Send(conn net.Conn, req Request) error {
	if req.SRC == "" {
		req.SRC = s.SRC
	}
	if req.DST == "" {
		return errors.New("Destination not provided")
	}
	if req.TOKEN == "" {
		if tkn, err := os.ReadFile(s.TOKENS[req.DST]); err == nil {
			req.TOKEN = string(tkn)
		}
	}
	return writeFramed(conn, req, s.LOGGER)
}
