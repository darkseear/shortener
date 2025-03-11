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

type Storage interface {
	ShortenURL(longURL string, userID string) (string, int)
	GetOriginalURL(shortURL string, userID string) (string, error)
	GetOriginalURLByUserID(cfg *config.Config, userID string) ([]models.URLPair, error)
	DeleteURLByUserID(shortURL []string, cfg *config.Config, userID string) error
	CreateTableDB(ctx context.Context) error
}

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
		return services.NewMemoryStorage(config), nil
	}

	logger.Log.Info("Create storage Memory")
	return services.NewMemoryStorage(config), nil
}
