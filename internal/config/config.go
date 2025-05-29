package config

import (
	"flag"
	"os"
	"sync"
)

// Config структура конфигурации приложения.
type Config struct {
	Address     string `env:"SERVER_ADDRESS"`
	URL         string `env:"BASE_URL"`
	LogLevel    string `env:"LOG_LEVEL"`
	MemoryFile  string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	SecretKey   string `env:"SECRET_KEY"`
	EnableHTTPS bool   `env:"ENABLE_HTTPS"`
	PprofAddr   string `env:"PPROF_ADDR"`
}

var (
	once            sync.Once
	flagAddress     string
	flagURL         string
	flagLogLevel    string
	flagFile        string
	flagDSN         string
	flagSecretKey   string
	flagEnableHTTPS bool
	flagPprofAddr   string
)

// registerFlags инициализирует флаги один раз.
func registerFlags() {
	once.Do(func() {
		flag.StringVar(&flagAddress, "a", "localhost:8080", "Server address")
		flag.StringVar(&flagURL, "b", "http://localhost:8080", "Base URL")
		flag.StringVar(&flagLogLevel, "l", "info", "Log level")
		flag.StringVar(&flagFile, "f", "memory.log", "File storage path")
		flag.StringVar(&flagDSN, "d", "", "Database DSN")
		flag.StringVar(&flagSecretKey, "sk", "secretkey", "Secret key for JWT")
		flag.BoolVar(&flagEnableHTTPS, "s", false, "Enable HTTPS (default: false)")
		flag.StringVar(&flagPprofAddr, "p", ":8081", "Address for pprof server")
	})
}

// New - создаёт конфиг с учётом флагов, переменных окружения и значений по умолчанию.
func New() *Config {
	registerFlags()

	if !flag.Parsed() {
		flag.Parse()
	}

	cfg := &Config{
		Address:     flagAddress,
		URL:         flagURL,
		LogLevel:    flagLogLevel,
		MemoryFile:  flagFile,
		DatabaseDSN: flagDSN,
		SecretKey:   flagSecretKey,
		EnableHTTPS: flagEnableHTTPS,
		PprofAddr:   flagPprofAddr,
	}

	// Переопределение значений переменными окружения
	setFromEnv(cfg)

	return cfg
}

// setFromEnv - обновляет конфиг значениями из переменных окружения.
func setFromEnv(cfg *Config) {
	envVars := map[string]*string{
		"SERVER_ADDRESS":    &cfg.Address,
		"BASE_URL":          &cfg.URL,
		"LOG_LEVEL":         &cfg.LogLevel,
		"FILE_STORAGE_PATH": &cfg.MemoryFile,
		"DATABASE_DSN":      &cfg.DatabaseDSN,
		"SECRET_KEY":        &cfg.SecretKey,
		"PPROF_ADDR":        &cfg.PprofAddr,
	}

	for env, ptr := range envVars {
		if val, ok := os.LookupEnv(env); ok {
			*ptr = val
		}
	}

	if val, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		if val == "true" || val == "1" {
			cfg.EnableHTTPS = true
		} else {
			cfg.EnableHTTPS = false
		}
	}
}

// DatabaseDSN: "host=localhost user=postgres password=1234567890 dbname=shorten sslmode=disable"
