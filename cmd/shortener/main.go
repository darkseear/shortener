package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/gzip"
	"github.com/darkseear/shortener/internal/handlers"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/proto"
	"github.com/darkseear/shortener/internal/services"
	"github.com/darkseear/shortener/internal/storage"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// HTTPServer - структура, представляющая HTTP сервер с роутером и конфигурацией.
type HTTPServer struct {
	Server *http.Server
	Router http.Handler
	Cfg    *config.Config
}

// GRPCServer - структура для gRPC сервера.
type GRPCServer struct {
	Server *grpc.Server
}

// App - основная структура приложения, содержащая серверы, хранилище и конфигурацию.
type App struct {
	HTTPServer *HTTPServer
	GRPCServer *GRPCServer
	Storage    storage.Storage
	Cfg        *config.Config
}

// newApp - инициализирует приложение, настраивает логирование, хранилище и роутер.
func newApp(ctx context.Context) (*App, error) {
	cfg := config.New()
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return nil, err
	}
	defer logger.Log.Sync()
	buildInfo()

	stor, err := storage.New(cfg)
	if err != nil {
		logger.Log.Error("Error store created")
		return nil, err
	}

	if cfg.MemoryFile != "" {
		absPath, err := filepath.Abs(cfg.MemoryFile)
		if err != nil {
			return nil, err
		}
		logger.Log.Info("absolute path memory file")
		cfg.MemoryFile = absPath
	}

	if cfg.DatabaseDSN != "" {
		stor.CreateTableDB(ctx)
	}

	router := logger.WhithLogging(gzip.GzipMiddleware(handlers.Routers(cfg, stor).Handle))
	httpSrv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	// Настройка gRPC сервера
	auth := services.NewAuthService(cfg.SecretKey).UnaryAuthInterceptor()
	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(auth))
	nss := proto.NewGRPCShortenerServer(stor, cfg)
	proto.RegisterSortenerServer(grpcSrv, nss)

	return &App{
		HTTPServer: &HTTPServer{
			Server: httpSrv,
			Router: router,
			Cfg:    cfg,
		},
		GRPCServer: &GRPCServer{
			Server: grpcSrv,
		},
		Storage: stor,
		Cfg:     cfg,
	}, nil
}

// Run - запускает приложение, инициализирует сервер и обрабатывает сигналы завершения.
func (a *App) Run(ctx context.Context) {

	// Инициализация pprof для профилирования
	go func() {
		pprofAddr := a.Cfg.PprofAddr
		logger.Log.Info("Starting pprof server", zap.String("address", pprofAddr))
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			logger.Log.Error("Error starting pprof server", zap.Error(err))
		}
	}()

	var err error
	// Запуск HTTP сервера с поддержкой HTTPS, если включено
	go func() {
		if a.Cfg.EnableHTTPS {
			logger.Log.Info("Starting HTTPS server", zap.String("address", a.Cfg.Address))
			err = a.HTTPServer.Server.Serve(autocert.NewListener("example.com"))
		} else {
			logger.Log.Info("Starting HTTP server", zap.String("address", a.Cfg.Address))
			err = a.HTTPServer.Server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Error starting HTTP server", zap.Error(err))
			log.Fatalf("Error starting HTTP server: %v", err)
		}

	}()

	// Запуск gRPC сервера
	go func() {
		listener, err := net.Listen("tcp", a.Cfg.GRPCAddr)
		if err != nil {
			logger.Log.Error("Error starting gRPC server", zap.Error(err))
			log.Fatalf("Error starting gRPC server: %v", err)
		}
		logger.Log.Info("Starting GRPC server", zap.String("GRPCaddress", a.Cfg.GRPCAddr))
		err = a.GRPCServer.Server.Serve(listener)
		if err != nil {
			logger.Log.Error("Error starting gRPC server", zap.Error(err))
			log.Fatalf("Error starting gRPC server: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Log.Info("Received shutdown signal, shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := a.Close(shutdownCtx); err != nil {
		logger.Log.Error("Error during shutdown", zap.Error(err))
	} else {
		logger.Log.Info("Server shutdown gracefully")
	}

}

// Close - закрывает приложение, останавливает сервер и освобождает ресурсы.
func (a *App) Close(ctx context.Context) error {
	var errs []error

	// Останавливаем gRPC сервер
	if a.GRPCServer != nil && a.GRPCServer.Server != nil {
		a.GRPCServer.Server.GracefulStop()
		logger.Log.Info("gRPC server stopped")
	}

	// Останавливаем HTTP сервер
	if a.HTTPServer != nil && a.HTTPServer.Server != nil {
		if err := a.HTTPServer.Server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Error shutting down HTTP server", zap.Error(err))
			errs = append(errs, err)
		} else {
			logger.Log.Info("HTTP server stopped")
		}
	}

	// Закрываем storage
	if a.Storage != nil {
		if err := a.Storage.Close(); err != nil {
			logger.Log.Error("Error closing storage", zap.Error(err))
			errs = append(errs, err)
		} else {
			logger.Log.Info("Storage closed")
		}
	}

	// Синхронизируем логгер
	if err := logger.Log.Sync(); err != nil {
		logger.Log.Error("Error syncing logger", zap.Error(err))
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	logger.Log.Info("Application closed successfully")
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()
	app, err := newApp(ctx)
	if err != nil {
		logger.Log.Error("Error create app", zap.Error(err))
		log.Fatalf("Error app: %v", err)
	}
	app.Run(ctx)
	defer app.Close(ctx)

}

// buildInfo возвращает информацию о сборке приложения
func buildInfo() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
}
