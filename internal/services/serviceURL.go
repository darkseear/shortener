package services

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/darkseear/shortener/internal/storage"
)

const sizeURL = 8

type URLService interface {
	ShortenURL(longURL string) string
	GetOriginalURL(shortURL string) (string, error)
}

type Handler struct {
	service URLService
	baseURL string
}

func NewHandler(service URLService, baseURL string) *Handler {
	return &Handler{service: service, baseURL: baseURL}
}

func ShortenURL(longURL string) string {
	s := storage.NewStorageServise().Storage
	b := make([]byte, sizeURL)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	shortURL := base64.URLEncoding.EncodeToString(b)
	s[shortURL] = longURL
	return shortURL
}

func GetOriginalURL(shortURL string) (string, error) {
	s := storage.NewStorageServise().Storage
	return s[shortURL], nil
}
