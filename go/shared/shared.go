package shared

type LoggingConfig struct {
	PATH  string
	RULES int
	LEVEL string
	NAME  string
}

type Node struct {
	IP   string
	PORT int
	LOG  LoggingConfig
}
