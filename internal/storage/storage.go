package storage

type MemoryStorage struct {
	Memory map[string]string
}
type URLService interface {
	ShortenURL(longURL string, fileName string) string
	GetOriginalURL(shortURL string) (string, error)
}
