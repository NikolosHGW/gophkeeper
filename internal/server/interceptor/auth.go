package interceptor

import (
	"context"
	"strings"

	"github.com/NikolosHGW/goph-keeper/internal/contextkey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type tokenValidator interface {
	ValidateToken(tokenString string) (int, error)
}

type AuthInterceptor struct {
	tokenService  tokenValidator
	noAuthMethods map[string]bool
}

func NewAuthInterceptor(tokenService tokenValidator, noAuthMethods []string) *AuthInterceptor {
	m := make(map[string]bool)
	for _, method := range noAuthMethods {
		m[method] = true
	}
	return &AuthInterceptor{
		tokenService:  tokenService,
		noAuthMethods: m,
	}
}

func (ai *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		if ai.noAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		userID, err := ai.authorize(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, contextkey.UserIDKey, userID)

		return handler(ctx, req)
	}
}

func (ai *AuthInterceptor) authorize(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "метаданные не предоставлены")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return 0, status.Error(codes.Unauthenticated, "токен авторизации не предоставлен")
	}

	accessToken := values[0]
	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	userID, err := ai.tokenService.ValidateToken(accessToken)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "недействительный токен доступа")
	}

	return userID, nil
}
