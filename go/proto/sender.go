package proto

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"larsys/go/lib"
	"net"
	"os"
)

type Host struct {
	IP   string
	PORT int
}

func (h *Host) HostID() string {
	return fmt.Sprintf("%s:%d", h.IP, h.PORT)
}

type Request struct {
	SRC    string
	DST    string
	TOKEN  string
	ACTION Action
}

type Sender struct {
	LOGGER *lib.Logger
	SRC    string
	TOKENS map[string]string
	DST    map[string]Host
}

func NewSender(src string, tokenPath map[string]string, logger *lib.Logger, destination map[string]Host) *Sender {
	return &Sender{
		DST:    destination,
		LOGGER: logger,
		SRC:    src,
		TOKENS: tokenPath,
	}
}

func (s *Sender) AddHost(name, ip string, port int) {
	s.DST[name] = Host{
		IP:   ip,
		PORT: port,
	}
	if s.LOGGER != nil {
		s.LOGGER.Printf("+++ Added host %s: %s:%d", name, ip, port)
	}
}

func (s *Sender) RemoveHost(name string) {
	delete(s.DST, name)

	if s.LOGGER != nil {
		s.LOGGER.Printf("--- Removed host %s", name)
	}
}

func (s *Sender) GetHost(dst string) string {
	host := s.DST[dst]
	return host.HostID()
}

func (s *Sender) Send(req Request) (Response, error) {
	if req.SRC == "" {
		req.SRC = s.SRC
	}
	if req.DST == "" {
		return Response{}, errors.New("Destination not provided")
	}

	if req.TOKEN == "" {
		tokenPath := GetTokenPath(req.DST, true)
		if _, err := os.Stat(tokenPath); err == nil {
			if tkn, err := os.ReadFile(tokenPath); err == nil {
				req.TOKEN = string(tkn)
			} else {
				s.LOGGER.Println("failed to read token:", err)
			}
		} else {
			// Token file does not exist – proceed without a token
			s.LOGGER.Println("token file not found, proceeding without token")
		}
	}

	host := s.DST[req.DST]
	host_id := host.HostID()
	req.DST = host_id

	conn, err := net.Dial("tcp", req.DST)
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()

	err = writeFramed(conn, req, s.LOGGER)
	if err != nil {
		return Response{}, err
	}
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var resp Response
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			if s.LOGGER != nil {
				s.LOGGER.Println("There was an error parsing the response:", err)
			}
			return Response{}, err
		}
		return resp, nil
	}
	return Response{}, errors.New("Something went wrong parsing the repsonse or we didn't get one")
}

func (s *Sender) Ping(dst string) (Response, error) {
	return s.Send(Request{
		SRC: s.SRC,
		DST: dst,
		ACTION: Action{
			NAME: "ping",
		},
	})
}

func (s *Sender) Register(dst string) (Response, error) {
	resp, err := s.Send(Request{
		SRC: s.SRC,
		DST: dst,
		ACTION: Action{
			NAME: "register",
		},
	})
	if err == nil { // was: err != nil
		if params, ok := resp.PARAMS.(map[string]any); ok {
			if token, ok := params["TOKEN"].(string); ok {
				if err := SaveToken(dst, token); err != nil {
					if s.LOGGER != nil {
						s.LOGGER.Println("save token failed:", err)
					}
				} else {
					if s.LOGGER != nil {
						s.LOGGER.Println("Token saved for", resp.SRC)
					}
					s.TOKENS[dst] = token
				}
			}
		}
	}
	return resp, err
}

func (s *Sender) Revoke(dst string) (Response, error) {
	return s.Send(Request{
		SRC: s.SRC,
		DST: dst,
		ACTION: Action{
			NAME: "revoke",
		},
	})
}

func (s *Sender) PluginInstall(dst string, params PLUGIN) (Response, error) {
	return s.Send(Request{
		SRC: s.SRC,
		DST: dst,
		ACTION: Action{
			NAME:   "plugin/install",
			PARAMS: params,
		},
	})
}

func (s *Sender) PluginUninstall(dst string, params PLUGIN) (Response, error) {
	return s.Send(Request{
		SRC: s.SRC,
		ACTION: Action{
			NAME:   "plugin/uninstall",
			PARAMS: params,
		},
	})
}
