package proto

import (
	"errors"
	"os"
	"path/filepath"
)

type PLUGIN struct {
	NAME          string
	VERSION       string
	DAEMON        string // daemon name for calling the file
	REQUIRED_DIRS []string
	MEM           int
	CPU           int
	TASKS         int
}

func (p *PLUGIN) Verify() error {
	if p.VERSION == "" {
		return errors.New("No version provided")
	}
	if p.MEM > 50 {
		p.MEM = 50
	}
	if p.CPU > 30 {
		p.CPU = 30
	}
	for _, directory := range p.REQUIRED_DIRS {
		os.MkdirAll(filepath.Join(PLUGIN_DIR, p.NAME, p.VERSION, directory), 0o755)
	}
	return nil
}
