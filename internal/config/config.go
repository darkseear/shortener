package config

import (
	"encoding/json"
	"flag"
	"os"
	"sync"

	"github.com/darkseear/shortener/internal/logger"
	"go.uber.org/zap"
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
	ConfigFile  string `env:"CONFIG"`
}
type ConfigFile struct {
	Address     string `json:"address"`      // -a /SERVER_ADDRESS
	URL         string `json:"url"`          // -b /BASE_URL
	MemoryFile  string `json:"memory_file"`  // -f /FILE_STORAGE_PATH
	DatabaseDSN string `json:"database_dsn"` // -d /DATABASE_DSN
	EnableHTTPS bool   `json:"enable_https"` // -s /ENABLE_HTTPS
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
	flagConfigFile  string
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
		flag.StringVar(&flagConfigFile, "c", "", "Path to config file")
		flag.StringVar(&flagConfigFile, "config", "", "Path to config file")
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
		ConfigFile:  flagConfigFile,
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
		"CONFIG":            &cfg.ConfigFile,
	}

	configFile, err := cfg.configFormFile()
	if err != nil {
		logger.Log.Error("Error reading config file", zap.Error(err))
	}

	for env, ptr := range envVars {
		if val, ok := os.LookupEnv(env); ok {
			*ptr = val
		} else if *ptr == "" {
			// Если переменная окружения не задана и флаг не указан, берём из файла конфигурации
			switch env {
			case "SERVER_ADDRESS":
				if configFile.Address != "" {
					*ptr = configFile.Address
				}
			case "BASE_URL":
				if configFile.URL != "" {
					*ptr = configFile.URL
				}
			case "FILE_STORAGE_PATH":
				if configFile.MemoryFile != "" {
					*ptr = configFile.MemoryFile
				}
			case "DATABASE_DSN":
				if configFile.DatabaseDSN != "" {
					*ptr = configFile.DatabaseDSN
				}
			}
		}
	}

	if val, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		if val == "true" || val == "1" {
			cfg.EnableHTTPS = true
		} else {
			cfg.EnableHTTPS = false
		}
	} else if !flagEnableHTTPS && !cfg.EnableHTTPS {
		// Если переменная окружения не задана и флаг не указан, берём из файла конфигурации
		cfg.EnableHTTPS = configFile.EnableHTTPS
	}
}

func (c *Config) configFormFile() (ConfigFile, error) {
	var configFile ConfigFile
	var filePath string
	if c.ConfigFile != "" {
		filePath = c.ConfigFile
	} else if os.Getenv("CONFIG") != "" {
		filePath = os.Getenv("CONFIG")
	} else {
		filePath = ""
	}

	if filePath == "" {
		return configFile, nil
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return configFile, err
	}
	if err := json.Unmarshal(file, &configFile); err != nil {
		return configFile, err
	}

	return configFile, nil
}

// DatabaseDSN: "host=localhost user=postgres password=1234567890 dbname=shorten sslmode=disable"
