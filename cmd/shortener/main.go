package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/gzip"
	"github.com/darkseear/shortener/internal/handlers"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/storage"
	"github.com/darkseear/shortener/internal/tls"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
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

	buildInfoShortener()

	storeTwo, err := storage.New(config)
	if err != nil {
		logger.Log.Error("Error store created")
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

	if config.DatabaseDSN != "" {
		storeTwo.CreateTableDB(context.Background())
	}

	// Запуск отдельного HTTP-сервера для pprof
	go func() {
		pprofAddr := config.PprofAddr
		logger.Log.Info("Starting pprof server", zap.String("address", pprofAddr))
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			logger.Log.Error("Error starting pprof server", zap.Error(err))
		}
	}()

	r := logger.WhithLogging(gzip.GzipMiddleware((handlers.Routers(config, storeTwo).Handle)))
	logger.Log.Info("Running server", zap.String("address", config.Address))

	if _, err := os.Stat(tls.CrtFile); os.IsNotExist(err) {
		logger.Log.Info("Certificate files not found, generating new ones")
		if err := tls.GenerateCerts(); err != nil {
			logger.Log.Error("Error generating certificates", zap.Error(err))
		}
	}
	if config.EnableHTTPS {
		logger.Log.Info("Starting HTTPS server", zap.String("address", config.Address))
		err = http.ListenAndServeTLS(config.Address, tls.CrtFile, tls.KeyFile, r)
	} else {
		logger.Log.Info("Starting HTTP server", zap.String("address", config.Address))
		err = http.ListenAndServe(config.Address, r)
	}
	return err
}

// buildInfo возвращает информацию о сборке приложения
// Эта функция используется для отображения информации о версии, дате сборки и коммите.
func buildInfoShortener() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
}
