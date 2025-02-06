package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

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
		db, err := sql.Open("pgx", config.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		return &Store{lDB: &storage.DBStorage{
			Db: db,
		}}, nil
	} else if config.MemoryFile != "" {
		return &Store{lm: &storage.MemoryStorage{
			Memory: make(map[string]string),
		}}, nil
	} else {
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

func (s *Store) ShortenURL(longURL string, fileName string) string {
	shortURL := GenerateShortURL(sizeURL)

	s.lm.Memory[shortURL] = longURL
	logger.Log.Info("Add in storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))

	if fileName != "" {
		p, err := NewProducer(fileName)
		if err != nil {
			panic(err)
		}
		m := models.MemoryFile{ShortURL: shortURL, LongURL: longURL}
		p.WriteMemoryFile(&m)
		defer p.Close()
	}

	return shortURL
}

func (s *Store) GetOriginalURL(shortURL string) (string, error) {
	count, ok := s.lm.Memory[shortURL]
	if !ok {
		return "", fmt.Errorf("error short")
	}
	logger.Log.Info("Get url from storage", zap.String("shortURL", shortURL), zap.String("originalURL", count))
	return count, nil
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
