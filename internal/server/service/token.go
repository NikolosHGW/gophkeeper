package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
	"github.com/golang-jwt/jwt/v4"
)

const TokenExp = time.Hour * 5

type token struct {
	log       logger.CustomLogger
	secretKey string
}

// NewToken - конструктор создания токен-сервиса.
func NewToken(log logger.CustomLogger, secretKey string) *token {
	return &token{log: log, secretKey: secretKey}
}

// GenerateJWT - генерирует токен на основе секретного ключа.
func (t *token) GenerateJWT(user *entity.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, entity.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: user.ID,
	})

	if t.secretKey == "" {
		t.log.LogInfo("для создании подписи токена секретный ключ пустой", fmt.Errorf("пустой secretKey"))
		return "", helper.ErrInternalServer
	}
	tokenString, err := token.SignedString([]byte(t.secretKey))
	if err != nil {
		t.log.LogInfo("ошибки при создании подписи токена: ", err)
		return "", helper.ErrInternalServer
	}

	return tokenString, nil
}

// ValidateToken валидирует токен.
func (s *token) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &entity.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*entity.Claims); ok && token.Valid {
		if claims.UserID == 0 {
			s.log.LogInfo("UserID отсутствует в клеймах токена", errors.New("invalid token: missing UserID"))
			return 0, errors.New("недействительный токен")
		}
		return claims.UserID, nil
	}

	return 0, errors.New("недействительный токен")
}
