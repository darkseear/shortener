package main

import (
	"net/http"
	"path/filepath"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/gzip"
	"github.com/darkseear/shortener/internal/handlers"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/services"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// запуск сервера
func run() error {

	//config
	config := config.New()
	address := config.Address
	LogLevel := config.LogLevel
	fileName := config.MemoryFile
	DDB := config.DatabaseDSN

	if err := logger.Initialize(LogLevel); err != nil {
		return err
	}

	m := services.NewMemory()
	err := services.MemoryFileSave(fileName, m)
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return err
	}
	config.MemoryFile = absPath
	fileName = config.MemoryFile
	//router chi
	r := logger.WhithLogging(gzip.GzipMiddleware((handlers.Routers(config.URL, m, fileName, DDB).Handle)))

	logger.Log.Info("Running server", zap.String("address", address))
	return http.ListenAndServe(address, r)
}
