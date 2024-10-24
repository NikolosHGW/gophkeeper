package service

import (
	"testing"
	"time"

	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestToken_GenerateJWT_Success(t *testing.T) {
	mockLogger := &mockLogger{}
	secretKey := "supersecretkey"
	tokenService := NewToken(mockLogger, secretKey)

	user := &entity.User{
		ID: 1,
	}

	tokenString, err := tokenService.GenerateJWT(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	parsedToken, err := jwt.ParseWithClaims(tokenString, &entity.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, parsedToken)

	claims, ok := parsedToken.Claims.(*entity.Claims)
	assert.True(t, ok)
	assert.Equal(t, user.ID, claims.UserID)
}

func TestToken_GenerateJWT_EmptySecretKey(t *testing.T) {
	mockLogger := &mockLogger{}
	tokenService := NewToken(mockLogger, "")

	user := &entity.User{
		ID: 1,
	}

	tokenString, err := tokenService.GenerateJWT(user)

	assert.Empty(t, tokenString)
	assert.ErrorIs(t, err, helper.ErrInternalServer)
}

func TestToken_ValidateToken_Success(t *testing.T) {
	mockLogger := &mockLogger{}
	secretKey := "supersecretkey"
	tokenService := NewToken(mockLogger, secretKey)

	user := &entity.User{
		ID: 1,
	}

	tokenString, err := tokenService.GenerateJWT(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	userID, err := tokenService.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userID)
}

func TestToken_ValidateToken_InvalidSignature(t *testing.T) {
	mockLogger := &mockLogger{}
	secretKey := "supersecretkey"
	tokenService := NewToken(mockLogger, secretKey)

	user := &entity.User{
		ID: 1,
	}

	tokenString, err := tokenService.GenerateJWT(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	anotherSecretKey := "anothersecretkey"
	anotherTokenService := NewToken(mockLogger, anotherSecretKey)

	userID, err := anotherTokenService.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Equal(t, 0, userID)
}

func TestToken_ValidateToken_ExpiredToken(t *testing.T) {
	mockLogger := &mockLogger{}
	secretKey := "supersecretkey"
	tokenService := NewToken(mockLogger, secretKey)

	user := &entity.User{
		ID: 1,
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, entity.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // истек час назад
		},
		UserID: user.ID,
	})

	tokenString, err := expiredToken.SignedString([]byte(secretKey))
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	userID, err := tokenService.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
	assert.Equal(t, 0, userID)
}

func TestToken_ValidateToken_InvalidClaims(t *testing.T) {
	mockLogger := &mockLogger{}
	secretKey := "supersecretkey"
	tokenService := NewToken(mockLogger, secretKey)

	invalidClaimsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
	})

	tokenString, err := invalidClaimsToken.SignedString([]byte(secretKey))
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	userID, err := tokenService.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Equal(t, 0, userID)
}

func TestToken_ValidateToken_EmptyToken(t *testing.T) {
	mockLogger := &mockLogger{}
	secretKey := "supersecretkey"
	tokenService := NewToken(mockLogger, secretKey)

	tokenString := ""

	userID, err := tokenService.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Equal(t, 0, userID)
}
