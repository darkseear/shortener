package config

import (
	"flag"
	"os"
)

type Config struct {
	Address     string
	URL         string
	LogLevel    string
	MemoryFile  string
	DatabaseDSN string
}

func New() *Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "server url")
	flag.StringVar(&config.URL, "b", "http://localhost:8080", "last url")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.StringVar(&config.MemoryFile, "f", "memory.log", "path storage file")
	// flag.StringVar(&config.DatabaseDSN, "d", "host=localhost user=postgres password=1234567890 dbname=shorten sslmode=disable", "Database DSN")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Database DSN")

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
	if val, state := os.LookupEnv("DATABASE_DSN"); state {
		config.DatabaseDSN = val
	}

	return &config
}
