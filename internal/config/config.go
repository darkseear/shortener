package config

import (
	"flag"
	"os"
)

type Config struct {
	Address string
	URL     string
}

func New() *Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "server url")
	flag.StringVar(&config.URL, "b", "http://localhost:8080", "last url")

	flag.Parse()

	serverAddress := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")

	if serverAddress != "" {
		config.Address = serverAddress
	}
	if baseURL != "" {
		config.URL = baseURL
	}

	return &config
}
