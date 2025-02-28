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

// type LocalMemory struct {
// 	localMemory *storage.MemoryStorage
// }

// type LocalDB struct {
// 	localDB *storage.DBStorage
// }

// func NewDB(strDB string) (*LocalDB, error) {
// 	db, err := sql.Open("pgx", strDB)
// 	if err != nil {
// 		logger.Log.Info("Error create DB", zap.Error(err))
// 		return nil, err
// 	}
// 	return &LocalDB{&storage.DBStorage{
// 		DB: db,
// 	}}, nil
// }

// func NewMemory() *LocalMemory {
// 	logger.Log.Info("Create storage")
// 	return &LocalMemory{&storage.MemoryStorage{
// 		Memory: make(map[string]string),
// 	}}
// }

type Store struct {
	lm    *storage.MemoryStorage
	lDB   *storage.DBStorage
	lFile *storage.FileStore
}

func NewStore(config *config.Config) (*Store, error) {
	var store Store

	switch {
	case config.DatabaseDSN != "":
		logger.Log.Info("Create storage DB")
		db, err := sql.Open("pgx", config.DatabaseDSN)
		if err != nil {
			logger.Log.Error("Error create storage DB", zap.Error(err))
			return nil, err
		}
		store.lDB = &storage.DBStorage{DB: db}

	case config.MemoryFile != "":
		logger.Log.Info("Create storage MemoryFile")
		store.lFile = &storage.FileStore{File: config.MemoryFile}

	default:
		logger.Log.Info("Create storage Memory")
		store.lm = &storage.MemoryStorage{Memory: make(map[string]string)}
	}

	return &store, nil
}

func (s *Store) ShortenURL(longURL string, cfg *config.Config, userID string) (string, int) {
	shortURL := GenerateShortURL(sizeURL)

	if cfg.DatabaseDSN != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		query := "INSERT INTO urls (long, shorten, userid) VALUES ($1, $2, $3) ON CONFLICT (long) DO NOTHING RETURNING *;"

		r, err := s.lDB.DB.ExecContext(ctx, query, longURL, shortURL, userID)

		if err != nil {
			logger.Log.Error("Not create write in table", zap.Error(err))
		}
		in, err := r.RowsAffected()
		if err != nil {
			logger.Log.Error("Rows affected error", zap.Error(err))
		}
		if in == 0 {
			logger.Log.Error("Conflict long")
			query := "SELECT shorten FROM urls WHERE long = $1"
			s := s.lDB.DB.QueryRowContext(ctx, query, longURL)
			fmt.Println(s)
			var short string
			err = s.Scan(&short)
			if err != nil {
				logger.Log.Error("scan error", zap.Error(err))
				return "", http.StatusBadRequest
			}
			err = s.Err()
			if err != nil {
				logger.Log.Error("db error", zap.Error(err))
				return "", http.StatusBadRequest
			}
			logger.Log.Info("In db storage", zap.String("shortURL", short), zap.String("longURL", longURL), zap.String("userID", userID))
			return short, http.StatusConflict
		} else {
			logger.Log.Info("Add in db storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL), zap.String("userID", userID))
			return shortURL, http.StatusCreated
		}

	} else if cfg.MemoryFile != "" {
		p, err := NewProducer(cfg.MemoryFile)
		if err != nil {
			logger.Log.Error("producer error", zap.Error(err))
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

func (s *Store) GetOriginalURL(shortURL string, cfg *config.Config, userID string) (string, error) {
	if cfg.DatabaseDSN != "" {
		logger.Log.Info("start get long url db")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var URL string
		query := "SELECT long FROM urls WHERE shorten = $1"
		row := s.lDB.DB.QueryRowContext(ctx, query, shortURL)
		err := row.Scan(&URL)
		if err != nil {
			logger.Log.Error("GetURL scan error", zap.Error(err))
			return "", err
		}
		err = row.Err()
		if err != nil {
			logger.Log.Error("GetURL error", zap.Error(err))
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
			logger.Log.Error("read memory file error", zap.Error(err))
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

func (s *Store) GetOriginalURLByUserID(cfg *config.Config, userID string) ([]models.URLPair, error) {
	logger.Log.Info("start get long url db")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var urls []models.URLPair
	if userID != "" {
		query := "SELECT shorten, long FROM urls WHERE userid = $1"
		rows, err := s.lDB.DB.QueryContext(ctx, query, userID)
		if err != nil {
			logger.Log.Error("GetURL query error", zap.Error(err))
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var URL string
			var OURL string
			if err := rows.Scan(&OURL, &URL); err != nil {
				logger.Log.Error("GetURL scan error", zap.Error(err))
				return urls, err
			}
			urls = append(urls, models.URLPair{ShortURL: "http://" + cfg.Address + "/" + OURL, LongURL: URL})
		}
		if err := rows.Err(); err != nil {
			logger.Log.Error("GetURL rows error", zap.Error(err))
			return nil, err
		}
		fmt.Println(urls)
	}

	return urls, nil
}

func (s *Store) CreateTableDB(ctx context.Context) error {
	logger.Log.Info("Create table shorten")
	query := "CREATE TABLE IF NOT EXISTS urls (" +
		"id SERIAL PRIMARY KEY," +
		"long VARCHAR(255) NOT NULL UNIQUE," +
		"shorten VARCHAR(50) NOT NULL UNIQUE," +
		"userID VARCHAR(50));"
	result, err := s.lDB.DB.ExecContext(ctx, query)
	if err != nil {
		logger.Log.Error("Error created table", zap.Error(err))
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
		logger.Log.Error("Error created table", zap.Error(err))
		panic(err)
	}
	shortURL := base64.URLEncoding.EncodeToString(b)
	return shortURL
}
