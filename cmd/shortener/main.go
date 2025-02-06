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
	LogLevel := config.LogLevel
	if err := logger.Initialize(LogLevel); err != nil {
		return err
	}

	if config.MemoryFile != "" {
		absPath, err := filepath.Abs(config.MemoryFile)
		if err != nil {
			return err
		}
		logger.Log.Info("absolute path memory file")
		config.MemoryFile = absPath
	}

	store, err := services.NewStore(config)
	if err != nil {
		logger.Log.Error("Error store created")
		return err
	}

	// if DDB == "" && fileName != "" {

	// 	absPath, err := filepath.Abs(fileName)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	config.MemoryFile = absPath
	// 	fileName = config.MemoryFile
	// 	err = services.MemoryFileSave(fileName, m)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	//router chi
	r := logger.WhithLogging(gzip.GzipMiddleware((handlers.Routers(config, store).Handle)))

	logger.Log.Info("Running server", zap.String("address", config.Address))
	return http.ListenAndServe(config.Address, r)
}
