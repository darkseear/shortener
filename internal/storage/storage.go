package storage

import (
	"database/sql"
)

type URLService interface {
	ShortenURL(longURL string, fileName string) string
	GetOriginalURL(shortURL string) (string, error)
}

type MemoryStorage struct {
	Memory map[string]string
}

type FileStore struct {
	File string
}
type DBStorage struct {
	DB *sql.DB
}
