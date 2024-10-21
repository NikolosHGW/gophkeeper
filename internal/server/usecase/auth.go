package usecase

import (
	"context"
	"fmt"

	pb "github.com/NikolosHGW/goph-keeper/api/authpb"
	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
)

type authRepo interface {
	User(context.Context, string) (*entity.User, error)
	userRepo
}

type auth struct {
	tokenService tokenServicer
	authRepo     authRepo
}

// NewAuth - конструктор юзкейса регистрации пользователя.
func NewAuth(tokenService tokenServicer, authRepo authRepo) *auth {
	return &auth{
		authRepo:     authRepo,
		tokenService: tokenService,
	}
}

// Handle - авторизация пользователя.
func (r *auth) Handle(ctx context.Context, req *pb.LoginUserRequest) (string, error) {
	user, err := r.authRepo.User(ctx, req.Login)
	if err != nil {
		return "", helper.ErrInternalServer
	}

	token, err := r.tokenService.GenerateJWT(user)
	if err != nil {
		return "", fmt.Errorf("ошибка при генерации токена: %w", err)
	}

	return token, nil
}
