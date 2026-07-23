package lib

import (
	"log"
	"os"
)

type LoggingConfig struct {
	PATH  string
	RULES int
	LEVEL string
	NAME  string
}

type Logger struct {
	*log.Logger
	file *os.File
}

func GetLogger(log_path string, log_rules int, format int) *Logger {
	log_f, err := os.OpenFile(log_path, log_rules, 0o644)
	if err != nil {
		panic(err)
	}
	return &Logger{
		Logger: log.New(log_f, "", format),
		file:   log_f,
	}
}

func (l *Logger) Close() error {
	return l.file.Close()
}
