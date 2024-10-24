package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NikolosHGW/goph-keeper/api/authpb"
	"github.com/NikolosHGW/goph-keeper/api/registerpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockRegisterClient struct {
	RegisterUserFunc func(
		ctx context.Context, req *registerpb.RegisterUserRequest, opts ...grpc.CallOption,
	) (*registerpb.RegisterUserResponse, error)
}

func (m *MockRegisterClient) RegisterUser(
	ctx context.Context, req *registerpb.RegisterUserRequest, opts ...grpc.CallOption,
) (*registerpb.RegisterUserResponse, error) {
	return m.RegisterUserFunc(ctx, req, opts...)
}

type mockLogger struct{}

func (n *mockLogger) LogInfo(message string, err error) {}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		mockRegister  func() *MockRegisterClient
		expectedToken string
		expectedErr   error
	}{
		{
			name:     "Успешная регистрация",
			login:    "user1",
			password: "pass1",
			mockRegister: func() *MockRegisterClient {
				return &MockRegisterClient{
					RegisterUserFunc: func(
						ctx context.Context, req *registerpb.RegisterUserRequest, opts ...grpc.CallOption,
					) (*registerpb.RegisterUserResponse, error) {
						return &registerpb.RegisterUserResponse{BearerToken: "token123"}, nil
					},
				}
			},
			expectedToken: "token123",
			expectedErr:   nil,
		},
		{
			name:     "Регистрация с ошибкой от RegisterClient",
			login:    "user2",
			password: "pass2",
			mockRegister: func() *MockRegisterClient {
				return &MockRegisterClient{
					RegisterUserFunc: func(
						ctx context.Context, req *registerpb.RegisterUserRequest, opts ...grpc.CallOption,
					) (*registerpb.RegisterUserResponse, error) {
						return nil, errors.New("grpc error")
					},
				}
			},
			expectedToken: "",
			expectedErr:   errors.New("ошибка при регистрации: grpc error"),
		},
		{
			name:     "Регистрация с пустым логином",
			login:    "",
			password: "pass3",
			mockRegister: func() *MockRegisterClient {
				return &MockRegisterClient{
					RegisterUserFunc: func(
						ctx context.Context, req *registerpb.RegisterUserRequest, opts ...grpc.CallOption,
					) (*registerpb.RegisterUserResponse, error) {
						return &registerpb.RegisterUserResponse{BearerToken: "token456"}, nil
					},
				}
			},
			expectedToken: "token456",
			expectedErr:   nil,
		},
		{
			name:     "Регистрация с пустым паролем",
			login:    "user4",
			password: "",
			mockRegister: func() *MockRegisterClient {
				return &MockRegisterClient{
					RegisterUserFunc: func(
						ctx context.Context, req *registerpb.RegisterUserRequest, opts ...grpc.CallOption,
					) (*registerpb.RegisterUserResponse, error) {
						return &registerpb.RegisterUserResponse{BearerToken: "token789"}, nil
					},
				}
			},
			expectedToken: "token789",
			expectedErr:   nil,
		},
		{
			name:     "Регистрация с контекстом, отмененным",
			login:    "user5",
			password: "pass5",
			mockRegister: func() *MockRegisterClient {
				return &MockRegisterClient{
					RegisterUserFunc: func(
						ctx context.Context, req *registerpb.RegisterUserRequest, opts ...grpc.CallOption,
					) (*registerpb.RegisterUserResponse, error) {
						return nil, context.Canceled
					},
				}
			},
			expectedToken: "",
			expectedErr:   context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGRPCClient := &GRPCClient{
				RegisterClient: tt.mockRegister(),
			}

			noOpLogger := &mockLogger{}

			authSvc := NewAuthService(mockGRPCClient, noOpLogger)

			token, err := authSvc.Register(context.Background(), tt.login, tt.password)

			assert.Equal(t, tt.expectedToken, token)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) LoginUser(
	ctx context.Context, req *authpb.LoginUserRequest, opts ...grpc.CallOption,
) (*authpb.LoginUserResponse, error) {
	args := m.Called(ctx, req, opts)
	var resp *authpb.LoginUserResponse
	if r := args.Get(0); r != nil {
		resp = r.(*authpb.LoginUserResponse)
	}
	return resp, args.Error(1)
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		mockAuth      func() *MockAuthClient
		expectedToken string
		expectedErr   error
	}{
		{
			name:     "Успешный вход",
			login:    "user1",
			password: "pass1",
			mockAuth: func() *MockAuthClient {
				mockAuthClient := new(MockAuthClient)
				mockAuthClient.On("LoginUser", mock.Anything, &authpb.LoginUserRequest{
					Login:    "user1",
					Password: "pass1",
				}, mock.Anything).Return(&authpb.LoginUserResponse{
					BearerToken: "token123",
				}, nil)
				return mockAuthClient
			},
			expectedToken: "token123",
			expectedErr:   nil,
		},
		{
			name:     "Вход с ошибкой аутентификации",
			login:    "user2",
			password: "wrongpass",
			mockAuth: func() *MockAuthClient {
				mockAuthClient := new(MockAuthClient)
				mockAuthClient.On("LoginUser", mock.Anything, &authpb.LoginUserRequest{
					Login:    "user2",
					Password: "wrongpass",
				}, mock.Anything).Return(nil, errors.New("invalid credentials"))
				return mockAuthClient
			},
			expectedToken: "",
			expectedErr:   errors.New("invalid credentials"),
		},
		{
			name:     "Вход с пустым логином",
			login:    "",
			password: "pass3",
			mockAuth: func() *MockAuthClient {
				mockAuthClient := new(MockAuthClient)
				mockAuthClient.On("LoginUser", mock.Anything, &authpb.LoginUserRequest{
					Login:    "",
					Password: "pass3",
				}, mock.Anything).Return(&authpb.LoginUserResponse{
					BearerToken: "token456",
				}, nil)
				return mockAuthClient
			},
			expectedToken: "token456",
			expectedErr:   nil,
		},
		{
			name:     "Вход с пустым паролем",
			login:    "user4",
			password: "",
			mockAuth: func() *MockAuthClient {
				mockAuthClient := new(MockAuthClient)
				mockAuthClient.On("LoginUser", mock.Anything, &authpb.LoginUserRequest{
					Login:    "user4",
					Password: "",
				}, mock.Anything).Return(&authpb.LoginUserResponse{
					BearerToken: "token789",
				}, nil)
				return mockAuthClient
			},
			expectedToken: "token789",
			expectedErr:   nil,
		},
		{
			name:     "Вход с отмененным контекстом",
			login:    "user5",
			password: "pass5",
			mockAuth: func() *MockAuthClient {
				mockAuthClient := new(MockAuthClient)
				mockAuthClient.On("LoginUser", mock.Anything, &authpb.LoginUserRequest{
					Login:    "user5",
					Password: "pass5",
				}, mock.Anything).Return(nil, context.Canceled)
				return mockAuthClient
			},
			expectedToken: "",
			expectedErr:   context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка моков
			mockAuthClient := tt.mockAuth()

			mockGRPCClient := &GRPCClient{
				AuthClient: mockAuthClient,
				// RegisterClient можно оставить nil, так как он не используется в тесте Login
			}

			noOpLogger := &mockLogger{}

			authSvc := &authService{
				registerClient: nil, // Не используется в тесте Login
				authClient:     mockGRPCClient.AuthClient,
				logger:         noOpLogger,
			}

			// Выполнение метода Login
			token, err := authSvc.Login(context.Background(), tt.login, tt.password)

			// Проверка результатов
			assert.Equal(t, tt.expectedToken, token)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			// Проверка, что все ожидания моков были выполнены
			mockAuthClient.AssertExpectations(t)
		})
	}
}
