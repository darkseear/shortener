package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/gzip"
	"github.com/darkseear/shortener/internal/handlers"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/storage"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

type server struct {
	Server  *http.Server
	cfg     *config.Config
	Storage storage.Storage
}

func newServer(ctx context.Context) (*server, error) {
	var serv server
	//config
	config := config.New()
	LogLevel := config.LogLevel
	if err := logger.Initialize(LogLevel); err != nil {
		return nil, err
	}

	defer logger.Log.Sync()
	buildInfo()

	storeTwo, err := storage.New(config)
	if err != nil {
		logger.Log.Error("Error store created")
		return nil, err
	}

	if config.MemoryFile != "" {
		absPath, err := filepath.Abs(config.MemoryFile)
		if err != nil {
			return nil, err
		}
		logger.Log.Info("absolute path memory file")
		config.MemoryFile = absPath
	}

	if config.DatabaseDSN != "" {
		storeTwo.CreateTableDB(ctx)
	}

	r := logger.WhithLogging(gzip.GzipMiddleware((handlers.Routers(config, storeTwo).Handle)))
	logger.Log.Info("Running server", zap.String("address", config.Address))

	serv.initServer(config.Address, r, config, storeTwo)
	return &serv, nil
}

func (serv *server) initServer(addr string, handler http.Handler, cfg *config.Config, storage storage.Storage) {
	serv.Server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	serv.cfg = cfg
	serv.Storage = storage
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	serv, err := newServer(ctx)
	if err != nil {
		logger.Log.Error("Error create server", zap.Error(err))
		log.Fatalf("Error server: %v", err)
	}
	defer serv.Close(ctx)
	if err := serv.run(ctx); err != nil {
		panic(err)
	}
}

// запуск сервера
func (serv *server) run(ctx context.Context) error {
	var err error

	go func() {
		<-ctx.Done()
		logger.Log.Info("Received shutdown signal, shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := serv.Close(shutdownCtx); err != nil {
			logger.Log.Error("Error during shutdown", zap.Error(err))
		} else {
			logger.Log.Info("Server shutdown gracefully")
		}
	}()

	// Запуск отдельного HTTP-сервера для pprof
	go func() {
		pprofAddr := serv.cfg.PprofAddr
		logger.Log.Info("Starting pprof server", zap.String("address", pprofAddr))
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			logger.Log.Error("Error starting pprof server", zap.Error(err))
		}
	}()

	if serv.cfg.EnableHTTPS {
		logger.Log.Info("Starting HTTPS server", zap.String("address", serv.cfg.Address))
		err = serv.Server.Serve(autocert.NewListener("example.com"))
	} else {
		logger.Log.Info("Starting HTTP server", zap.String("address", serv.cfg.Address))
		err = serv.Server.ListenAndServe()
	}
	return err
}

func (serv *server) Close(ctx context.Context) error {
	logger.Log.Info("Closing storage and stopping server")
	// Закрытие сервера
	if err := serv.Server.Shutdown(ctx); err != nil {
		logger.Log.Error("Error shutting down server", zap.Error(err))
		return err
	}
	if err := serv.Storage.Close(); err != nil {
		logger.Log.Error("Error closing storage", zap.Error(err))
		return err
	}
	if err := logger.Log.Sync(); err != nil {
		logger.Log.Error("Error sync logger", zap.Error(err))
		return err
	}
	logger.Log.Info("closed successfully")
	return nil
}

// buildInfo возвращает информацию о сборке приложения
// Эта функция используется для отображения информации о версии, дате сборки и коммите.
func buildInfo() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
}
