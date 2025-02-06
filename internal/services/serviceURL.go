package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/models"
	"github.com/darkseear/shortener/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const sizeURL int64 = 8

type LocalMemory struct {
	localMemory *storage.MemoryStorage
}

type LocalDB struct {
	localDB *storage.DBStorage
}

type Store struct {
	lm  *storage.MemoryStorage
	lDB *storage.DBStorage
}

func NewStore(config *config.Config) (*Store, error) {
	if config.DatabaseDSN != "" {
		logger.Log.Info("Create storage DB")
		db, err := sql.Open("pgx", config.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		return &Store{lDB: &storage.DBStorage{
			Db: db,
		}}, nil
	} else if config.MemoryFile != "" {
		logger.Log.Info("Create storage MemoryFile")
		return &Store{lm: &storage.MemoryStorage{
			Memory: make(map[string]string),
		}}, nil
	} else {
		logger.Log.Info("Create storage Memory")
		return &Store{lm: &storage.MemoryStorage{
			Memory: make(map[string]string),
		}}, nil
	}
}

func NewDB(strDB string) (*LocalDB, error) {
	db, err := sql.Open("pgx", strDB)
	if err != nil {
		return nil, err
	}
	return &LocalDB{&storage.DBStorage{
		Db: db,
	}}, nil
}

func NewMemory() *LocalMemory {
	logger.Log.Info("Create storage")
	return &LocalMemory{&storage.MemoryStorage{
		Memory: make(map[string]string),
	}}
}

func (s *Store) ShortenURL(longURL string, cfg *config.Config) string {
	shortURL := GenerateShortURL(sizeURL)

	if cfg.DatabaseDSN != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		r, err := s.lDB.Db.ExecContext(ctx, "INSERT INTO urls (long, shorten) SELECT $1, $2 WHERE NOT EXISTS (SELECT 1 FROM urls WHERE shorten = $3)", longURL, shortURL, shortURL)
		if err != nil {
			logger.Log.Error("Not create write in table")
		}
		logger.Log.Info("Add in db storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
		fmt.Println(r)
		return shortURL
	} else if cfg.MemoryFile != "" {
		p, err := NewProducer(cfg.MemoryFile)
		if err != nil {
			panic(err)
		}
		m := models.MemoryFile{ShortURL: shortURL, LongURL: longURL}
		p.WriteMemoryFile(&m)
		logger.Log.Info("Add in file storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
		defer p.Close()
		return shortURL
	} else {
		s.lm.Memory[shortURL] = longURL
		logger.Log.Info("Add in memory storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
		return shortURL
	}
}

func (s *Store) GetOriginalURL(shortURL string, cfg *config.Config) (string, error) {
	if cfg.DatabaseDSN != "" {
		logger.Log.Info("start get long url db")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var URL string

		row := s.lDB.Db.QueryRowContext(ctx, "SELECT long FROM urls WHERE shorten = $1", shortURL)
		err := row.Scan(&URL)
		if err != nil {
			logger.Log.Error("GetURL scan error")
			return "", err
		}
		err = row.Err()
		if err != nil {
			logger.Log.Error("GetURL error")
			return "", err
		}

		return URL, nil
	} else if cfg.MemoryFile != "" {
		logger.Log.Info("start get long url memory file")
		c, err := NewConsumer(cfg.MemoryFile)
		if err != nil {
			panic(err)
		}
		list, err := c.ReadMemoryFileAll()
		if err != nil {
			panic(err)
		}
		count, ok := list[shortURL]
		if !ok {
			return "", fmt.Errorf("error short")
		}
		return count, nil
	} else {
		logger.Log.Info("start get long url memory")
		count, ok := s.lm.Memory[shortURL]
		if !ok {
			return "", fmt.Errorf("error short")
		}
		logger.Log.Info("Get url from storage", zap.String("shortURL", shortURL), zap.String("originalURL", count))
		return count, nil
	}
}

func (s *Store) CreateTableDB(ctx context.Context) error {
	logger.Log.Info("Create table shorten")
	result, err := s.lDB.Db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS urls ("+
		"id SERIAL PRIMARY KEY,"+
		"long VARCHAR(255) NOT NULL,"+
		"shorten VARCHAR(50) NOT NULL UNIQUE)")
	if err != nil {
		logger.Log.Error("Error created table")
		return err
	}
	fmt.Println(result)
	logger.Log.Info("Created table")
	return nil
}

func GenerateShortURL(i int64) string {
	b := make([]byte, i)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	shortURL := base64.URLEncoding.EncodeToString(b)
	return shortURL
}
