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
)

// sizeURL - размер короткого адреса.
const sizeURL int64 = 8

// MemoryStorage - структура для хранения в памяти.
// Используется для тестирования и в случае, если не требуется постоянное хранение данных.
type MemoryStorage struct {
	Memory map[string]string
	cfg    *config.Config
}

// NewMemoryStorage - конструктор для создания нового экземпляра MemoryStorage.
// Принимает конфигурацию в качестве параметра и инициализирует память.
func NewMemoryStorage(cfg *config.Config) *MemoryStorage {
	return &MemoryStorage{
		Memory: make(map[string]string), cfg: cfg}
}

// GetOriginalURL - метод для получения оригинального URL по короткому адресу.
// Принимает короткий адрес и идентификатор пользователя в качестве параметров.
func (m *MemoryStorage) GetOriginalURL(shortURL string, userID string) (string, error) {
	logger.Log.Info("start get long url memory")
	count, ok := m.Memory[shortURL]
	if !ok {
		return "", fmt.Errorf("error short")
	}
	logger.Log.Info("Get url from storage", zap.String("shortURL", shortURL), zap.String("originalURL", count))
	return count, nil
}

// ShortenURL - метод для сокращения URL.
// Принимает длинный URL и идентификатор пользователя в качестве параметров.
func (m *MemoryStorage) ShortenURL(longURL string, userID string) (string, int) {
	shortURL := GenerateShortURL(sizeURL)
	m.Memory[shortURL] = longURL
	logger.Log.Info("Add in memory storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
	return shortURL, http.StatusCreated
}

// CreateTableDB - метод для создания таблицы в базе данных.
// В данном случае он не реализован, так как используется только в памяти.
func (m *MemoryStorage) CreateTableDB(ctx context.Context) error {
	panic("unimplemented")
}

// DeleteURLByUserID - метод для удаления URL по идентификатору пользователя.
// В данном случае он не реализован, так как используется только в памяти.
func (m *MemoryStorage) DeleteURLByUserID(shortURL []string, userID string) error {
	panic("unimplemented")
}

// GetOriginalURLByUserID - метод для получения оригинального URL по идентификатору пользователя.
// В данном случае он не реализован, так как используется только в памяти.
func (m *MemoryStorage) GetOriginalURLByUserID(userID string) ([]models.URLPair, error) {
	panic("unimplemented")
}

//
//end memory

// db
// DBStorage - структура для работы с базой данных.
type DBStorage struct {
	DB  *sql.DB
	cfg *config.Config
}

// NewDBStorage - конструктор для создания нового экземпляра DBStorage.
// Принимает указатель на базу данных и конфигурацию в качестве параметров.
func NewDBStorage(db *sql.DB, cfg *config.Config) *DBStorage {
	return &DBStorage{DB: db, cfg: cfg}
}

// GetOriginalURL - метод для получения оригинального URL по короткому адресу.
// Принимает короткий адрес и идентификатор пользователя в качестве параметров.
func (d *DBStorage) GetOriginalURL(shortURL string, userID string) (string, error) {
	logger.Log.Info("start get long url db")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var DBUrlShorten = &models.DBUrlShorten{}
	query := "SELECT shorten, long, userid, is_deleted FROM urls WHERE shorten = $1"
	rows, err := d.DB.QueryContext(ctx, query, shortURL)
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
}

// / ShortenURL - метод для сокращения URL.
// Принимает длинный URL и идентификатор пользователя в качестве параметров.
func (d *DBStorage) ShortenURL(longURL string, userID string) (string, int) {
	shortURL := GenerateShortURL(sizeURL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := "INSERT INTO urls (long, shorten, userid) VALUES ($1, $2, $3) RETURNING *;"
	_, err := d.DB.ExecContext(ctx, query, longURL, shortURL, userID)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			logger.Log.Error("Conflict long")
			query := "SELECT shorten FROM urls WHERE long = $1"
			s := d.DB.QueryRowContext(ctx, query, longURL)
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
}

// GetOriginalURLByUserID - метод для получения оригинального URL по идентификатору пользователя.
// Принимает идентификатор пользователя в качестве параметра.
func (d *DBStorage) GetOriginalURLByUserID(userID string) ([]models.URLPair, error) {
	logger.Log.Info("start get long url db")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var urls []models.URLPair
	if userID != "" {
		query := "SELECT shorten, long FROM urls WHERE userid = $1"
		rows, err := d.DB.QueryContext(ctx, query, userID)
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
			urls = append(urls, models.URLPair{ShortURL: "http://" + d.cfg.Address + "/" + OURL, LongURL: URL})
		}
		if err := rows.Err(); err != nil {
			logger.Log.Error("GetURL rows error", zap.Error(err))
			return nil, err
		}
	}

	return urls, nil
}

// DeleteURLByUserID - метод для удаления URL по идентификатору пользователя.
// Принимает короткий адрес и идентификатор пользователя в качестве параметров.
func (d *DBStorage) DeleteURLByUserID(shortURL []string, userID string) error {
	logger.Log.Info("start delete url")
	if d.cfg.DatabaseDSN != "" {
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
		_, err := d.DB.ExecContext(ctx, query, pq.Array(shortURL), userID)
		if err != nil {
			logger.Log.Error("DeleteURL error", zap.Error(err))
			return err
		}
		return nil
	}
	return nil
}

// CreateTableDB - метод для создания таблицы в базе данных.
// Принимает контекст в качестве параметра.
// Создает таблицу urls, если она не существует.
func (d *DBStorage) CreateTableDB(ctx context.Context) error {
	logger.Log.Info("Create table shorten")
	query := "CREATE TABLE IF NOT EXISTS urls (" +
		"id SERIAL PRIMARY KEY," +
		"long VARCHAR(255) NOT NULL UNIQUE," +
		"shorten VARCHAR(50) NOT NULL UNIQUE," +
		"userID VARCHAR(50)," +
		"is_deleted BOOL DEFAULT false);"
	_, err := d.DB.ExecContext(ctx, query)
	if err != nil {
		logger.Log.Error("Error created table", zap.Error(err))
		return err
	}
	logger.Log.Info("Created table")
	return nil
}

//end db

// file
// FileStore - структура для работы с файловым хранилищем.
// Используется для хранения данных в файле.
type FileStore struct {
	File string
	cfg  *config.Config
}

// NewFileStore - конструктор для создания нового экземпляра FileStore.
// Принимает путь к файлу и конфигурацию в качестве параметров.
func NewFileStore(file string, cfg *config.Config) *FileStore {
	return &FileStore{File: file, cfg: cfg}
}

// GetOriginalURL - метод для получения оригинального URL по короткому адресу.
// Принимает короткий адрес и идентификатор пользователя в качестве параметров.
func (f *FileStore) GetOriginalURL(shortURL string, userID string) (string, error) {
	logger.Log.Info("start get long url memory file")
	c, err := NewConsumer(f.cfg.MemoryFile)
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
}

// ShortenURL - метод для сокращения URL.
// Принимает длинный URL и идентификатор пользователя в качестве параметров.
// Генерирует короткий адрес и записывает его в файл.
func (f *FileStore) ShortenURL(longURL string, userID string) (string, int) {
	shortURL := GenerateShortURL(sizeURL)
	p, err := NewProducer(f.cfg.MemoryFile)
	if err != nil {
		logger.Log.Error("producer error", zap.Error(err))
		panic(err)
	}
	m := models.MemoryFile{ShortURL: shortURL, LongURL: longURL}
	p.WriteMemoryFile(&m)
	logger.Log.Info("Add in file storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
	defer p.Close()
	return shortURL, http.StatusCreated
}

// CreateTableDB - метод для создания таблицы в файловом хранилище.
// В данном случае он не реализован, так как используется только в памяти.
func (f *FileStore) CreateTableDB(ctx context.Context) error {
	panic("unimplemented")
}

// DeleteURLByUserID - метод для удаления URL по идентификатору пользователя.
// Принимает короткий адрес и идентификатор пользователя в качестве параметров.
func (f *FileStore) DeleteURLByUserID(shortURL []string, userID string) error {
	panic("unimplemented")
}

// GetOriginalURLByUserID - метод для получения оригинального URL по идентификатору пользователя.
// Принимает идентификатор пользователя в качестве параметра.
func (f *FileStore) GetOriginalURLByUserID(userID string) ([]models.URLPair, error) {
	panic("unimplemented")
}

//
//end file

// генератор короткого адреса, возможно надо сосздать папку хелперсов и туда его
// GenerateShortURL - функция для генерации короткого адреса.
// Принимает длину короткого адреса в качестве параметра и возвращает сгенерированный адрес.
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
