package storage

import "database/sql"

type MemoryStorage struct {
	Memory map[string]string
}

type FileStore struct {
	File string
}
type DBStorage struct {
	Db *sql.DB
}

type URLService interface {
	ShortenURL(longURL string, fileName string) string
	GetOriginalURL(shortURL string) (string, error)
}
