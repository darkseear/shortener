package config

import (
	"flag"
	"os"
)

type Config struct {
	Address    string
	URL        string
	LogLevel   string
	MemoryFile string
}

func New() *Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "server url")
	flag.StringVar(&config.URL, "b", "http://localhost:8080", "last url")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.MemoryFile, "f", "memory.log", "path storage file")

	flag.Parse()

	if val, state := os.LookupEnv("SERVER_ADDRESS"); state {
		config.Address = val
	}
	if val, state := os.LookupEnv("BASE_URL"); state {
		config.URL = val
	}
	if val, state := os.LookupEnv("LOG_LEVEL"); state {
		config.LogLevel = val
	}
	if val, state := os.LookupEnv("FILE_STORAGE_PATH"); state {
		config.MemoryFile = val
	}

	return &config
}
