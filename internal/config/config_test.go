package config

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// Сброс состояния флагов перед тестом
	defer func() {
		flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
	}()

	cfg := New()
	fmt.Printf("Running  %s, %s, %s", cfg.Address, cfg.URL, cfg.MemoryFile)
	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.Equal(t, "http://localhost:8080", cfg.URL)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "memory.log", cfg.MemoryFile)
	assert.Equal(t, "", cfg.DatabaseDSN)
	assert.Equal(t, "secretkey", cfg.SecretKey)
}

func TestConfigWithEnv(t *testing.T) {
	// Сброс состояния флагов и переменных окружения
	defer func() {
		flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("SECRET_KEY")
	}()

	// Установка переменных окружения для теста
	os.Setenv("SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost:9090")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("FILE_STORAGE_PATH", "test_memory.log")
	os.Setenv("DATABASE_DSN", "test_dsn")
	os.Setenv("SECRET_KEY", "test_secret")

	cfg := New()

	assert.Equal(t, "localhost:9090", cfg.Address)
	assert.Equal(t, "http://localhost:9090", cfg.URL)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "test_memory.log", cfg.MemoryFile)
	assert.Equal(t, "test_dsn", cfg.DatabaseDSN)
	assert.Equal(t, "test_secret", cfg.SecretKey)
}
