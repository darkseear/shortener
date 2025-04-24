package storage

import (
	"context"
	"database/sql"

	"go.uber.org/zap"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/models"
	"github.com/darkseear/shortener/internal/services"
)

// Storage - интерфейс для работы с хранилищем.
type Storage interface {
	ShortenURL(longURL string, userID string) (string, int)
	GetOriginalURL(shortURL string, userID string) (string, error)
	GetOriginalURLByUserID(userID string) ([]models.URLPair, error)
	DeleteURLByUserID(shortURL []string, userID string) error
	CreateTableDB(ctx context.Context) error
}

// New - функция для создания нового хранилища.
// В зависимости от конфигурации создается либо хранилище в памяти, либо в файле, либо в базе данных.
func New(config *config.Config) (Storage, error) {

	if config.DatabaseDSN != "" {
		logger.Log.Info("Create storage DB")
		db, err := sql.Open("pgx", config.DatabaseDSN)
		if err != nil {
			logger.Log.Error("Error create storage DB", zap.Error(err))
			return nil, err
		}
		return services.NewDBStorage(db, config), nil
	}
	if config.MemoryFile != "" {
		logger.Log.Info("Create storage MemoryFile")
		return services.NewFileStore(config.MemoryFile, config), nil
	}

	logger.Log.Info("Create storage Memory")
	return services.NewMemoryStorage(config), nil
}
