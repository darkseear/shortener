package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
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
			DB: db,
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
		DB: db,
	}}, nil
}

func NewMemory() *LocalMemory {
	logger.Log.Info("Create storage")
	return &LocalMemory{&storage.MemoryStorage{
		Memory: make(map[string]string),
	}}
}

func (s *Store) ShortenURL(longURL string, cfg *config.Config) (string, int) {
	shortURL := GenerateShortURL(sizeURL)

	if cfg.DatabaseDSN != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		r, err := s.lDB.DB.ExecContext(ctx, "INSERT INTO urls (long, shorten) VALUES ($1, $2) ON CONFLICT (long) DO NOTHING RETURNING *;", longURL, shortURL)
		if err != nil {
			fmt.Println(err.Error())
			logger.Log.Error("Not create write in table")
		}
		in, err := r.RowsAffected()
		if err != nil {
			fmt.Println(err.Error())
		}
		if in == 0 {
			logger.Log.Error("Conflict long")
			s := s.lDB.DB.QueryRowContext(ctx, "SELECT shorten FROM urls WHERE long = $1", longURL)
			fmt.Println(s)
			var short string
			err = s.Scan(&short)
			if err != nil {
				logger.Log.Error("scan error")
				return "", http.StatusBadRequest
			}
			err = s.Err()
			if err != nil {
				logger.Log.Error("error")
				return "", http.StatusBadRequest
			}
			logger.Log.Info("In db storage", zap.String("shortURL", short), zap.String("longURL", longURL))
			return short, http.StatusConflict
		} else {
			logger.Log.Info("Add in db storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
			return shortURL, http.StatusCreated
		}

	} else if cfg.MemoryFile != "" {
		p, err := NewProducer(cfg.MemoryFile)
		if err != nil {
			panic(err)
		}
		m := models.MemoryFile{ShortURL: shortURL, LongURL: longURL}
		p.WriteMemoryFile(&m)
		logger.Log.Info("Add in file storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
		defer p.Close()
		return shortURL, http.StatusCreated
	} else {
		s.lm.Memory[shortURL] = longURL
		logger.Log.Info("Add in memory storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
		return shortURL, http.StatusCreated
	}
}

func (s *Store) GetOriginalURL(shortURL string, cfg *config.Config) (string, error) {
	if cfg.DatabaseDSN != "" {
		logger.Log.Info("start get long url db")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var URL string

		row := s.lDB.DB.QueryRowContext(ctx, "SELECT long FROM urls WHERE shorten = $1", shortURL)
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
	result, err := s.lDB.DB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS urls ("+
		"id SERIAL PRIMARY KEY,"+
		"long VARCHAR(255) NOT NULL UNIQUE,"+
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
