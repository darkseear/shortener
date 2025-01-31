package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/storage"
	"go.uber.org/zap"
)

const sizeURL int64 = 8

type LocalMemory struct {
	localMemory *storage.MemoryStorage
}

func NewMemory() *LocalMemory {
	logger.Log.Info("Create storage")
	return &LocalMemory{&storage.MemoryStorage{
		Memory: make(map[string]string),
	}}
}

func (s *LocalMemory) ShortenURL(longURL string) string {
	shortURL := GenerateShortURL(sizeURL)
	s.localMemory.Memory[shortURL] = longURL
	logger.Log.Info("Add in storage", zap.String("shortURL", shortURL), zap.String("longURL", longURL))
	return shortURL
}

func (s *LocalMemory) GetOriginalURL(shortURL string) (string, error) {
	count, ok := s.localMemory.Memory[shortURL]
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
