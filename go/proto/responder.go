package proto

import (
	"errors"
	"log"
	"net"
)

type Response struct {
	SRC    string
	STATUS int
	MSG    string
	PARAMS any
}

type Responder struct {
	LOGGER *log.Logger
	SRC    string
}

func NewResponder(src string, logger *log.Logger) *Responder {
	return &Responder{
		LOGGER: logger,
		SRC:    src,
	}
}

func (r *Responder) Respond(conn net.Conn, resp Response) error {
	if resp.SRC == "" {
		resp.SRC = r.SRC
	}
	return writeFramed(conn, resp, r.LOGGER)
}

func (r *Responder) OK(conn net.Conn, msg string) error {
	if msg == "" {
		msg = "OK"
	}
	resp := Response{
		SRC:    r.SRC,
		STATUS: 0,
		MSG:    msg,
		PARAMS: struct{}{},
	}
	return writeFramed(conn, resp, r.LOGGER)
}

func (r *Responder) Error(conn net.Conn, msg string, err error) error {
	if msg == "" {
		msg = "An error ocurred"
	}
	if err == nil {
		return errors.New("No error message provided")
	}
	return r.Respond(conn, Response{
		STATUS: 1,
		MSG:    msg,
		PARAMS: struct{ ERROR error }{ERROR: err}},
	)
}

func (r *Responder) Unauthorised(conn net.Conn) error {
	resp := Response{
		SRC:    r.SRC,
		STATUS: 1,
		MSG:    "Unauthorised request",
		PARAMS: struct{}{},
	}
	return writeFramed(conn, resp, r.LOGGER)
}
func (r *Responder) AlreadyRegistered(conn net.Conn) error {
	resp := Response{
		SRC:    r.SRC,
		STATUS: 1,
		MSG:    "Client already registered",
		PARAMS: struct{}{},
	}
	return writeFramed(conn, resp, r.LOGGER)
}

func (r *Responder) Register(conn net.Conn, token string) error {
	resp := Response{
		SRC:    r.SRC,
		STATUS: 0,
		MSG:    "+++ Client Registered +++",
		PARAMS: struct{ TOKEN string }{
			TOKEN: token,
		},
	}
	return writeFramed(conn, resp, r.LOGGER)
}
