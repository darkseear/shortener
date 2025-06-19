package proto

import (
	"context"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/storage"
)

type GRPCShortenerServer struct {
	UnimplementedSortenerServer
	Store *storage.Storage
	Cfg   *config.Config
}

func NewGRPCShortenerServer(store *storage.Storage, cfg *config.Config) *GRPCShortenerServer {
	return &GRPCShortenerServer{
		Store: store,
		Cfg:   cfg,
	}
}

func (s *GRPCShortenerServer) AddURL(ctx context.Context, req *AddURLRequest) (*AddURLResponse, error) {
	// Реализация метода для добавления URL
	return nil, nil
}

func (s *GRPCShortenerServer) GetURL(ctx context.Context, req *GetURLRequest) (*GetURLResponse, error) {
	// Реализация метода для получения оригинального URL по короткому
	return nil, nil
}

func (s *GRPCShortenerServer) DeleteURL(ctx context.Context, req *DeleteURLRequest) (*DeleteURLResponse, error) {
	// Реализация метода для удаления URL
	return nil, nil
}

func (s *GRPCShortenerServer) ListURL(ctx context.Context, req *ListURLRequest) (*ListURLResponse, error) {
	// Реализация метода для получения списка всех URL
	return nil, nil
}
func (s *GRPCShortenerServer) PingDB(ctx context.Context, req *PingDBRequest) (*PingDBResponse, error) {
	// Реализация метода для проверки соединения с базой данных
	return nil, nil
}

func (s *GRPCShortenerServer) Stats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	// Реализация метода для получения статистики
	return nil, nil
}
func (s *GRPCShortenerServer) ShortenBatch(ctx context.Context, req *ShortenBatchRequest) (*ShortenBatchResponse, error) {
	// Реализация метода для пакетного сокращения URL
	return nil, nil
}

func (s *GRPCShortenerServer) Shorten(ctx context.Context, req *ShortenRequest) (*ShortenResponse, error) {
	// Реализация метода для сокращения URL
	return nil, nil
}
