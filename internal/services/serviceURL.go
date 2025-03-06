package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/models"
	"github.com/darkseear/shortener/internal/storage"
)

const sizeURL int64 = 8

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
		query := "INSERT INTO urls (long, shorten, userid) VALUES ($1, $2, $3) RETURNING *;"
		_, err := s.lDB.DB.ExecContext(ctx, query, longURL, shortURL, userID)

		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
				logger.Log.Error("Conflict long")
				query := "SELECT shorten FROM urls WHERE long = $1"
				s := s.lDB.DB.QueryRowContext(ctx, query, longURL)
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
			}
			logger.Log.Error("Not create write in table", zap.Error(err))
			return "", http.StatusBadRequest
		}
		logger.Log.Info("Add in db storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL), zap.String("userID", userID))
		return shortURL, http.StatusCreated

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

		var DBUrlShorten = &models.DBUrlShorten{}
		query := "SELECT shorten, long, userid, is_deleted FROM urls WHERE shorten = $1"
		rows, err := s.lDB.DB.QueryContext(ctx, query, shortURL)
		if err != nil {
			logger.Log.Error("GetURL query error", zap.Error(err))
			return "", err
		}
		defer rows.Close()
		if rows.Next() {
			if err = rows.Scan(&DBUrlShorten.ShortURL, &DBUrlShorten.LongURL, &DBUrlShorten.UserID, &DBUrlShorten.DeletedFlag); err != nil {
				logger.Log.Error("GetURL scan error", zap.Error(err))
				return "", err
			}
		}

		if DBUrlShorten.DeletedFlag {
			logger.Log.Error("GetURL error, url is deleted", zap.Error(err))
			return "GoneStatus", nil
		}
		err = rows.Err()
		if err != nil {
			logger.Log.Error("GetURL error", zap.Error(err))
			return "", err
		}

		return DBUrlShorten.LongURL, nil
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
	}

	return urls, nil
}

func (s *Store) DeleteURLByUserID(shortURL []string, cfg *config.Config, userID string) error {
	logger.Log.Info("start delete url")
	if cfg.DatabaseDSN != "" {
		logger.Log.Info("start delete url db")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		query := `
		UPDATE urls 
		SET is_deleted = true 
		WHERE 
		shorten = ANY($1) 
		AND 
		userID = $2;`
		_, err := s.lDB.DB.ExecContext(ctx, query, pq.Array(shortURL), userID)
		if err != nil {
			logger.Log.Error("DeleteURL error", zap.Error(err))
			return err
		}
		return nil
	}
	return nil
}

func (s *Store) CreateTableDB(ctx context.Context) error {
	logger.Log.Info("Create table shorten")
	query := "CREATE TABLE IF NOT EXISTS urls (" +
		"id SERIAL PRIMARY KEY," +
		"long VARCHAR(255) NOT NULL UNIQUE," +
		"shorten VARCHAR(50) NOT NULL UNIQUE," +
		"userID VARCHAR(50)," +
		"is_deleted BOOL DEFAULT false);"
	_, err := s.lDB.DB.ExecContext(ctx, query)
	if err != nil {
		logger.Log.Error("Error created table", zap.Error(err))
		return err
	}
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
