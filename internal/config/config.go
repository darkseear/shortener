package config

import "flag"

type Config struct {
	Address string
	URL     string
}

func New() *Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "server url")
	flag.StringVar(&config.URL, "b", "http://localhost:8080", "last url")

	flag.Parse()
	return &config
}