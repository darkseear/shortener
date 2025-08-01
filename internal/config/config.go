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
	Address       string `env:"SERVER_ADDRESS"`
	URL           string `env:"BASE_URL"`
	LogLevel      string `env:"LOG_LEVEL"`
	MemoryFile    string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN   string `env:"DATABASE_DSN"`
	SecretKey     string `env:"SECRET_KEY"`
	EnableHTTPS   bool   `env:"ENABLE_HTTPS"`
	PprofAddr     string `env:"PPROF_ADDR"`
	ConfigFile    string `env:"CONFIG"`
	TrustedSubnet string `env:"TRUSTED_SUBNET"`
	GRPCAddr      string `env:"GRPC_ADDR"` // Адрес gRPC сервера
}

// ConfigFile структура для хранения конфигурации из файла.
// Используется для загрузки параметров из JSON-файла конфигурации.
type ConfigFile struct {
	Address       string `json:"address"`        // -a /SERVER_ADDRESS
	URL           string `json:"url"`            // -b /BASE_URL
	MemoryFile    string `json:"memory_file"`    // -f /FILE_STORAGE_PATH
	DatabaseDSN   string `json:"database_dsn"`   // -d /DATABASE_DSN
	EnableHTTPS   bool   `json:"enable_https"`   // -s /ENABLE_HTTPS
	TrustedSubnet string `json:"trusted_subnet"` // -t /TRUSTED_SUBNET
	GRPCAddr      string `json:"grpc_addr"`      // Адрес gRPC сервера
}

var (
	once              sync.Once
	flagAddress       string
	flagURL           string
	flagLogLevel      string
	flagFile          string
	flagDSN           string
	flagSecretKey     string
	flagEnableHTTPS   bool
	flagPprofAddr     string
	flagConfigFile    string
	flagTrustedSubnet string
	flagGRPCAddr      string
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
		flag.StringVar(&flagTrustedSubnet, "t", "", "Trusted subnet for internal requests")
		flag.StringVar(&flagGRPCAddr, "g", "localhost:9090", "gRPC server address")
	})
}

// New - создаёт конфиг с учётом флагов, переменных окружения и значений по умолчанию.
func New() *Config {
	registerFlags()

	if !flag.Parsed() {
		flag.Parse()
	}

	cfg := &Config{
		Address:       flagAddress,
		URL:           flagURL,
		LogLevel:      flagLogLevel,
		MemoryFile:    flagFile,
		DatabaseDSN:   flagDSN,
		SecretKey:     flagSecretKey,
		EnableHTTPS:   flagEnableHTTPS,
		PprofAddr:     flagPprofAddr,
		TrustedSubnet: flagTrustedSubnet,
		ConfigFile:    flagConfigFile,
		GRPCAddr:      flagGRPCAddr,
	}

	// Переопределение значений переменными окружения
	setFromEnv(cfg)

	return cfg
}

// setFromEnv - обновляет конфиг значениями из переменных окружения.
func setFromEnv(cfg *Config) {
	configFile := getConfigFile(cfg)

	setStringFields(cfg, configFile)
	setEnableHTTPS(cfg, configFile)
}

// getConfigFile - конфиг из файла.
func getConfigFile(cfg *Config) ConfigFile {
	configFile, err := cfg.configFormFile()
	if err != nil {
		logger.Log.Error("Error reading config file", zap.Error(err))
	}
	return configFile
}

// setStringFields - строки файла.
func setStringFields(cfg *Config, configFile ConfigFile) {
	envVars := map[string]*string{
		"SERVER_ADDRESS":    &cfg.Address,
		"BASE_URL":          &cfg.URL,
		"LOG_LEVEL":         &cfg.LogLevel,
		"FILE_STORAGE_PATH": &cfg.MemoryFile,
		"DATABASE_DSN":      &cfg.DatabaseDSN,
		"SECRET_KEY":        &cfg.SecretKey,
		"PPROF_ADDR":        &cfg.PprofAddr,
		"CONFIG":            &cfg.ConfigFile,
		"TRUSTED_SUBNET":    &cfg.TrustedSubnet,
		"GRPC_ADDR":         &cfg.GRPCAddr,
	}

	for env, ptr := range envVars {
		if val, ok := os.LookupEnv(env); ok {
			*ptr = val
		} else if *ptr == "" {
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
			case "TRUSTED_SUBNET":
				if configFile.TrustedSubnet != "" {
					*ptr = configFile.TrustedSubnet
				}
			case "GRPC_ADDR":
				if configFile.GRPCAddr != "" {
					*ptr = configFile.GRPCAddr
				}
			}
		}
	}
}

//setEnableHTTPS - устанавливает значение EnableHTTPS из переменной окружения или из файла конфигурации.

func setEnableHTTPS(cfg *Config, configFile ConfigFile) {
	if val, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		cfg.EnableHTTPS = val == "true" || val == "1"
	} else if !flagEnableHTTPS && !cfg.EnableHTTPS {
		cfg.EnableHTTPS = configFile.EnableHTTPS
	}
}

// configFormFile читает конфигурацию из файла, если указан путь к файлу.
// Если файл не указан, возвращает пустую структуру ConfigFile.
// Если файл указан, но не может быть прочитан или распарсен, возвращает ошибку.

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
