package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/darkseear/shortener/internal/logger"
)

// contextKey - тип для ключей в контексте.
type contextKey string

// AuthService - структура для работы с авторизацией.
type AuthService struct {
	secretKey string
}

// NewAuthService - конструктор для создания нового AuthService.
func NewAuthService(secretKey string) *AuthService {
	return &AuthService{secretKey: secretKey}
}

// GenerateToken - метод для генерации JWT токена.
// Он принимает userID и возвращает сгенерированный токен или ошибку.
func (s *AuthService) GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken - метод для проверки JWT токена.
// Он принимает строку токена и возвращает userID, если токен действителен, или ошибку.
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["userID"].(string)
		return userID, nil
	}

	return "", errors.New("invalid token")
}

// SetCookie - метод для установки куки с токеном.
// Он принимает http.ResponseWriter и userID, генерирует токен и устанавливает его в куки.
func (s *AuthService) SetCookie(w http.ResponseWriter, userID string) string {
	tokenString, err := s.GenerateToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(72 * time.Hour),
		HttpOnly: true,
	})
	return userID
}

// IssueCookie - метод для проверки наличия куки и его валидности.
// Если куки нет или он не валиден, то генерируется новый токен и устанавливается в куки.
func (s *AuthService) IssueCookie(w http.ResponseWriter, r *http.Request, userID string) string {
	cookie, err := r.Cookie("auth_token")
	if err != nil || cookie == nil {
		logger.Log.Info("Создаем и применяем новое куки если его нет")
		UID := s.SetCookie(w, userID)
		return UID
	}

	userID, err = s.ValidateToken(cookie.Value)
	if err != nil {
		logger.Log.Info("Токен не действителен")
		UID := s.SetCookie(w, userID)
		return UID
	}

	return userID
}

// UnaryAuthInterceptor возвращает grpc.UnaryServerInterceptor для проверки JWT токена.
func (s *AuthService) UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var token string
		var err error
		var userID string
		var newCtx context.Context

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		} else {
			values := md.Get("auth_token")
			if len(values) > 0 {
				token = values[0]
			} else {
				userID = fmt.Sprintf("%d", int(math.Floor(1000+math.Floor(9000*rand.Float64()))))
				token, err = s.GenerateToken(userID)
				if err != nil {
					logger.Log.Error("Ошибка при генерации токена", zap.Error(err))
					return nil, status.Error(codes.Internal, "failed to generate token")
				}
			}
		}

		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		userID, err = s.ValidateToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		clientIP := ""
		if mdClientIP, ok := md["client_ip"]; ok && len(mdClientIP) > 0 {
			clientIP = mdClientIP[0]
		}

		// Добавляем в context для дальнейшего использования
		newCtx = context.WithValue(ctx, contextKey("userID"), userID)
		newCtx = context.WithValue(newCtx, contextKey("client_ip"), clientIP)
		newCtx = context.WithValue(newCtx, contextKey("auth_token"), token)
		// Добавляем userID и auth_token в метаданные gRPC запроса
		md = metadata.Pairs("userID", userID, "auth_token", token, "client_ip", clientIP)
		newCtx = metadata.NewIncomingContext(newCtx, md)
		// Передаем новый контекст с userID дальше в цепочку вызовов
		logger.Log.Info("Проверка токена прошла успешно")
		res, err := handler(newCtx, req)
		if err != nil {
			logger.Log.Error("Ошибка при обработке запроса", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}

		if err := grpc.SendHeader(newCtx, md); err != nil {
			logger.Log.Error("Ошибка при отправке заголовков", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to send headers")
		}
		return res, nil
	}
}

// GetUserIDFromContext - извлекает userID из контекста запроса.
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return "", errors.New("userID not found in context")
	}
	return userID, nil
}

// GetUserIDFromMetadata - извлекает userID из метаданных gRPC запроса.
func GetUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}

	values := md["userID"]
	if len(values) == 0 {
		return "", errors.New("userID not found in metadata")
	}

	userID := values[0]
	if userID == "" {
		return "", errors.New("userID is empty")
	}

	return userID, nil
}

// GetClientIPFromMetadata - извлекает IP клиента из метаданных gRPC запроса.
func GetClientIPFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}

	values := md["client_ip"]
	if len(values) == 0 {
		return "", errors.New("client_ip not found in metadata")
	}

	clientIP := values[0]
	if clientIP == "" {
		return "", errors.New("client_ip is empty")
	}

	return clientIP, nil
}
