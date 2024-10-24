package handler

import (
	"context"

	pb "github.com/NikolosHGW/goph-keeper/api/authpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type auth interface {
	Handle(context.Context, *pb.LoginUserRequest) (string, error)
}

// AuthServer - структура gRPC сервера для авторизации пользователя.
type AuthServer struct {
	pb.UnimplementedAuthServer

	authUseCase auth
}

// NewAuthServer - конструктор gRPC сервера для авторизации пользователя.
func NewAuthServer(authUseCase auth) *AuthServer {
	return &AuthServer{authUseCase: authUseCase}
}

// LoginUser - реализация RPC сервиса.
func (s *AuthServer) LoginUser(
	ctx context.Context,
	req *pb.LoginUserRequest,
) (*pb.LoginUserResponse, error) {
	err := validateLoginPasswordRequest(req.Login, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "неправильный запрос: %v", err)
	}

	token, err := s.authUseCase.Handle(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка при авторизации: %v", err)
	}

	return &pb.LoginUserResponse{
		BearerToken: token,
	}, nil
}
