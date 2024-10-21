package service

import (
	"context"
	"fmt"

	"github.com/NikolosHGW/goph-keeper/api/authpb"
	"github.com/NikolosHGW/goph-keeper/api/registerpb"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
)

type authService struct {
	registerClient registerpb.RegisterClient
	authClient     authpb.AuthClient
	logger         logger.CustomLogger
}

func NewAuthService(grpcClient *GRPCClient, logger logger.CustomLogger) *authService {
	return &authService{
		registerClient: grpcClient.RegisterClient,
		authClient:     grpcClient.AuthClient,
		logger:         logger,
	}
}

func (s *authService) Register(ctx context.Context, login, password string) (string, error) {
	req := &registerpb.RegisterUserRequest{
		Login:    login,
		Password: password,
	}
	resp, err := s.registerClient.RegisterUser(ctx, req)
	if err != nil {
		s.logger.LogInfo("Ошибка регистрации", err)
		return "", fmt.Errorf("ошибка при регистрации: %w", err)
	}
	return resp.BearerToken, nil
}

func (s *authService) Login(ctx context.Context, login, password string) (string, error) {
	req := &authpb.LoginUserRequest{
		Login:    login,
		Password: password,
	}
	res, err := s.authClient.LoginUser(ctx, req)
	if err != nil {
		return "", fmt.Errorf("ошибка при логине: %w", err)
	}
	return res.BearerToken, nil
}
