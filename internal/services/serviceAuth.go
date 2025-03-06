package services

import (
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/darkseear/shortener/internal/logger"
)

type AuthService struct {
	secretKey string
}

func NewAuthService(secretKey string) *AuthService {
	return &AuthService{secretKey: secretKey}
}

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
