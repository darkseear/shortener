package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Log - синглтон для логирования.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует логгер с заданным уровнем логирования.
//
// Уровень логирования может быть "debug", "info", "warn", "error", "dpanic", "panic" или "fatal".
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

// Write - реализуем интерфейс http.ResponseWriter, захват данных ответа.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

// WriteHeader - реализуем интерфейс http.ResponseWriter, захват статуса.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// WhithLogging - обертка для http.Handler, которая добавляет логирование запросов и ответов.
func WhithLogging(h http.Handler) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		uri := zap.String("uri", r.RequestURI)
		method := zap.String("method", r.Method)

		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)

		Log.Info("request HTTP",
			uri,
			method,
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size),
			zap.Int("status", responseData.status),
		)
	}

	return http.HandlerFunc(logFn)
}
