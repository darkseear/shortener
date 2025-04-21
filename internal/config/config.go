package config

import (
	"flag"
	"os"
)

type Config struct {
	Address     string `env:"SERVER_ADDRESS"`
	URL         string `env:"BASE_URL"`
	LogLevel    string `env:"LOG_LEVEL"`
	MemoryFile  string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	SecretKey   string `env:"SECRET_KEY"`
}

func New() *Config {
	config := Config{
		Address:     "localhost:8080",
		URL:         "http://localhost:8080",
		LogLevel:    "info",
		MemoryFile:  "memory.log",
		DatabaseDSN: "host=localhost user=postgres password=1234567890 dbname=shorten sslmode=disable",
		SecretKey:   "secretkey",
	}

	flag.StringVar(&config.Address, "a", "", "server url")
	flag.StringVar(&config.URL, "b", "", "last url")
	flag.StringVar(&config.LogLevel, "l", "", "log level")
	flag.StringVar(&config.MemoryFile, "f", "", "path storage file")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Database DSN")
	flag.StringVar(&config.SecretKey, "s", "", "Key for JWT")

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
	if val, state := os.LookupEnv("SECRET_KEY"); state {
		config.SecretKey = val
	}

	return &config
}
