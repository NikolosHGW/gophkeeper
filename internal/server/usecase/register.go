package usecase

import (
	"context"
	"fmt"

	pb "github.com/NikolosHGW/goph-keeper/api/registerpb"
	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
)

type userRepo interface {
	Save(context.Context, *entity.User) error
	ExistsByLogin(context.Context, string) (bool, error)
}

type registerServicer interface {
	CreateUser(*pb.RegisterUserRequest) (*entity.User, error)
}

type tokenServicer interface {
	GenerateJWT(*entity.User) (string, error)
}

type register struct {
	registerService registerServicer
	tokenService    tokenServicer
	userRepo        userRepo
}

// NewRegister - конструктор юзкейса регистрации пользователя.
func NewRegister(registerService registerServicer, tokenService tokenServicer, userRepo userRepo) *register {
	return &register{
		registerService: registerService,
		userRepo:        userRepo,
		tokenService:    tokenService,
	}
}

// Handle - регистрация пользователя.
func (r *register) Handle(ctx context.Context, req *pb.RegisterUserRequest) (string, error) {
	isLoginExist, err := r.userRepo.ExistsByLogin(ctx, req.Login)
	if err != nil {
		return "", helper.ErrInternalServer
	}
	if isLoginExist {
		return "", helper.ErrLoginAlreadyExists
	}

	user, err := r.registerService.CreateUser(req)
	if err != nil {
		return "", fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	if err := r.userRepo.Save(ctx, user); err != nil {
		return "", fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	token, err := r.tokenService.GenerateJWT(user)
	if err != nil {
		return "", fmt.Errorf("ошибка при генерации токена: %w", err)
	}

	return token, nil
}
