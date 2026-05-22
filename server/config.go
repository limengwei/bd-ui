package server

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host    string
	Port    int
	RootDir string
	URL     string
}

func GetConfig() *Config {
	rootDir, _ := os.Getwd()

	port := 3000
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			port = n
		}
	}

	host := "127.0.0.1"
	if v := os.Getenv("HOST"); v != "" {
		host = v
	}

	return &Config{
		Host:    host,
		Port:    port,
		RootDir: rootDir,
		URL:     fmt.Sprintf("http://%s:%d", host, port),
	}
}
