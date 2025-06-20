package proto

import (
	"context"
	"database/sql"
	"net"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/services"
	"github.com/darkseear/shortener/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCShortenerServer - структура, представляющая gRPC сервер для сокращения URL.
type GRPCShortenerServer struct {
	UnimplementedSortenerServer
	Store storage.Storage
	Cfg   *config.Config
}

// NewGRPCShortenerServer - конструктор для создания нового gRPC сервера.
func NewGRPCShortenerServer(store storage.Storage, cfg *config.Config) *GRPCShortenerServer {
	return &GRPCShortenerServer{
		Store: store,
		Cfg:   cfg,
	}
}

// AddURL - метод для добавления нового URL в систему.
func (s *GRPCShortenerServer) AddURL(ctx context.Context, req *AddURLRequest) (*AddURLResponse, error) {

	userID, err := services.GetUserIDFromMetadata(ctx)
	if err != nil || userID == "" {
		logger.Log.Error("failed to control user ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "user ID is not provided")
	}

	if req == nil || req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "Empty url")
	}

	short, storeStatus := s.Store.ShortenURL(req.Url, userID)
	if storeStatus != 201 && storeStatus != 200 {
		return nil, status.Error(codes.Internal, "failed to shorten URL")
	}

	return &AddURLResponse{
		ShortUrl: s.Cfg.URL + "/" + short,
	}, nil
}

// GetURL - метод для получения оригинального URL по короткому.
func (s *GRPCShortenerServer) GetURL(ctx context.Context, req *GetURLRequest) (*GetURLResponse, error) {
	userID, err := services.GetUserIDFromMetadata(ctx)
	if err != nil || userID == "" {
		logger.Log.Error("failed to control user ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "user ID is not provided")
	}

	paramURLID := req.GetShortUrl()
	if paramURLID == "" {
		return nil, status.Error(codes.InvalidArgument, "short url is empty")
	}

	originalURL, err := s.Store.GetOriginalURL(paramURLID, userID)
	if err == nil && originalURL == "GoneStatus" {
		return nil, status.Error(codes.NotFound, "url is deleted")
	}
	if err != nil {
		return nil, status.Error(codes.NotFound, "url not found")
	}

	return &GetURLResponse{
		OriginalUrl: originalURL,
	}, nil
}

// DeleteURL - метод для удаления URL по короткому идентификатору.
func (s *GRPCShortenerServer) DeleteURL(ctx context.Context, req *DeleteURLRequest) (*DeleteURLResponse, error) {
	userID, err := services.GetUserIDFromMetadata(ctx)
	if err != nil || userID == "" {
		logger.Log.Error("failed to control user ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "user ID is not provided")
	}

	if req == nil || len(req.ShortUrls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no urls provided")
	}

	err = s.Store.DeleteURLByUserID(req.ShortUrls, userID)
	if err != nil {
		logger.Log.Error("Delete error", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete urls")
	}

	logger.Log.Info("User", zap.String("Delete url is userID:", userID))
	return &DeleteURLResponse{Success: true}, nil
}

// ListURL - метод для получения списка всех URL, добавленных пользователем.
func (s *GRPCShortenerServer) ListURL(ctx context.Context, req *ListURLRequest) (*ListURLResponse, error) {
	userID, err := services.GetUserIDFromMetadata(ctx)
	if err != nil || userID == "" {
		logger.Log.Error("failed to control user ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "user ID is not provided")
	}

	urls, err := s.Store.GetOriginalURLByUserID(userID)
	if err != nil {
		logger.Log.Error("failed to get URLs by user ID", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get URLs")
	}
	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "no URLs found")
	}

	var items []*URLItem
	for _, u := range urls {
		items = append(items, &URLItem{
			UserId:      userID,
			ShortUrl:    s.Cfg.URL + "/" + u.ShortURL,
			OriginalUrl: u.LongURL,
		})
	}

	logger.Log.Info("User", zap.String("userID", userID))
	return &ListURLResponse{Urls: items}, nil
}

// PingDB - метод для проверки доступности базы данных.
func (s *GRPCShortenerServer) PingDB(ctx context.Context, req *PingDBRequest) (*PingDBResponse, error) {
	db, err := sql.Open("pgx", s.Cfg.DatabaseDSN)
	if err != nil {
		logger.Log.Error("failed to open database", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to open database")
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		logger.Log.Error("failed to ping database", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "database is unavailable")
	}

	return &PingDBResponse{Ok: true}, nil
}

// Stats - метод для получения статистики по URL и пользователям.
func (s *GRPCShortenerServer) Stats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	// Проверка trusted_subnet
	if s.Cfg.TrustedSubnet == "" {
		return nil, status.Error(codes.PermissionDenied, "trusted subnet not configured")
	}

	clientIP, err := services.GetClientIPFromMetadata(ctx)
	if err != nil || clientIP == "" {
		return nil, status.Error(codes.PermissionDenied, "client IP not provided")
	}

	_, subnet, err := net.ParseCIDR(s.Cfg.TrustedSubnet)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "invalid trusted subnet")
	}

	ip := net.ParseIP(clientIP)
	if ip == nil || !subnet.Contains(ip) {
		return nil, status.Error(codes.PermissionDenied, "client IP not allowed")
	}

	stats, err := s.Store.Stats(ctx)
	if err != nil {
		logger.Log.Error("Error getting stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get stats")
	}
	if stats.URLs == 0 && stats.Users == 0 {
		return nil, status.Error(codes.NotFound, "no stats available")
	}

	logger.Log.Info("Stats requested", zap.Int("URLs", stats.URLs), zap.Int("Users", stats.Users))
	logger.Log.Info("Client IP", zap.String("IP", clientIP))

	return &StatsResponse{
		Urls:  int64(stats.URLs),
		Users: int64(stats.Users),
	}, nil
}

// ShortenBatch - метод для пакетного сокращения URL.
func (s *GRPCShortenerServer) ShortenBatch(ctx context.Context, req *ShortenBatchRequest) (*ShortenBatchResponse, error) {
	userID, err := services.GetUserIDFromMetadata(ctx)
	if err != nil || userID == "" {
		logger.Log.Error("failed to control user ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "user ID is not provided")
	}

	if req == nil || len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no urls provided")
	}

	var results []*ShortenBatchResponseItem
	for _, item := range req.Items {
		shortenURL, _ := s.Store.ShortenURL(item.OriginalUrl, userID)
		results = append(results, &ShortenBatchResponseItem{
			CorrelationId: item.CorrelationId,
			ShortUrl:      s.Cfg.URL + "/" + shortenURL,
		})
	}

	return &ShortenBatchResponse{
		Items: results,
	}, nil

}

// Shorten - метод для сокращения URL.
func (s *GRPCShortenerServer) Shorten(ctx context.Context, req *ShortenRequest) (*ShortenResponse, error) {
	userID, err := services.GetUserIDFromMetadata(ctx)
	if err != nil || userID == "" {
		logger.Log.Error("failed to control user ID", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "user ID is not provided")
	}

	longURL := req.GetUrl()
	if longURL == "" {
		return nil, status.Error(codes.InvalidArgument, "Empty url")
	}

	shortenURL, storeStatus := s.Store.ShortenURL(longURL, userID)
	if storeStatus != 201 && storeStatus != 200 {
		return nil, status.Error(codes.Internal, "failed to shorten URL")
	}

	return &ShortenResponse{
		ShortUrl: s.Cfg.URL + "/" + shortenURL,
	}, nil
}
