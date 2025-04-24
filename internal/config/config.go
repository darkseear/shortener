package config

import (
	"flag"
	"os"
)

// Config структура конфигурации приложения.
type Config struct {
	Address     string `env:"SERVER_ADDRESS"`
	URL         string `env:"BASE_URL"`
	LogLevel    string `env:"LOG_LEVEL"`
	MemoryFile  string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	SecretKey   string `env:"SECRET_KEY"`
}

// New - конструктор конфига (флаги , енв, дефолт)
func New() *Config {
	config := Config{
		Address:    "localhost:8080",
		URL:        "http://localhost:8080",
		LogLevel:   "info",
		MemoryFile: "memory.log",
		// DatabaseDSN: "host=localhost user=postgres password=1234567890 dbname=shorten sslmode=disable",
		DatabaseDSN: "",
		SecretKey:   "secretkey",
	}

	flag.StringVar(&config.Address, "a", config.Address, "server url")
	flag.StringVar(&config.URL, "b", config.URL, "last url")
	flag.StringVar(&config.LogLevel, "l", config.LogLevel, "log level")
	flag.StringVar(&config.MemoryFile, "f", config.MemoryFile, "path storage file")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "Database DSN")
	flag.StringVar(&config.SecretKey, "s", config.SecretKey, "Key for JWT")

	flag.Parse()

	envVars := map[string]*string{
		"SERVER_ADDRESS":    &config.Address,
		"BASE_URL":          &config.URL,
		"LOG_LEVEL":         &config.LogLevel,
		"FILE_STORAGE_PATH": &config.MemoryFile,
		"DATABASE_DSN":      &config.DatabaseDSN,
		"SECRET_KEY":        &config.SecretKey,
	}

	for key, ptr := range envVars {
		if val, state := os.LookupEnv(key); state {
			*ptr = val
		}
	}

	return &config
}
