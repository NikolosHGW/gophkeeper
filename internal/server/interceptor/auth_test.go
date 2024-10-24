package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/NikolosHGW/goph-keeper/internal/contextkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(tokenString string) (int, error) {
	args := m.Called(tokenString)
	return args.Int(0), args.Error(1)
}

func TestAuthInterceptor_Unary(t *testing.T) {
	mockValidator := new(MockTokenValidator)
	noAuthMethods := []string{"/package.Service/NoAuthMethod"}
	interceptor := NewAuthInterceptor(mockValidator, noAuthMethods)

	tests := []struct {
		name           string
		method         string
		metadata       metadata.MD
		mockSetup      func()
		expectedResult interface{}
		expectedError  error
		handler        grpc.UnaryHandler
	}{
		{
			name:   "Метод без аутентификации должен обходить проверку",
			method: "/package.Service/NoAuthMethod",
			metadata: metadata.New(map[string]string{
				"authorization": "Bearer validtoken",
			}),
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError:  nil,
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name:   "Успешная аутентификация",
			method: "/package.Service/AuthMethod",
			metadata: metadata.New(map[string]string{
				"authorization": "Bearer validtoken",
			}),
			mockSetup: func() {
				mockValidator.On("ValidateToken", "validtoken").Return(123, nil)
			},
			expectedResult: 123,
			expectedError:  nil,
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				userID, ok := ctx.Value(contextkey.UserIDKey).(int)
				if !ok {
					return nil, errors.New("userID не найден в контексте")
				}
				return userID, nil
			},
		},
		{
			name:           "Отсутствуют метаданные",
			method:         "/package.Service/AuthMethod",
			metadata:       nil,
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError:  status.Error(codes.Unauthenticated, "метаданные не предоставлены"),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name:     "Отсутствует токен авторизации",
			method:   "/package.Service/AuthMethod",
			metadata: metadata.New(map[string]string{
				// "authorization" отсутствует
			}),
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError:  status.Error(codes.Unauthenticated, "токен авторизации не предоставлен"),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name:   "Недействительный токен",
			method: "/package.Service/AuthMethod",
			metadata: metadata.New(map[string]string{
				"authorization": "Bearer invalidtoken",
			}),
			mockSetup: func() {
				mockValidator.On("ValidateToken", "invalidtoken").Return(0, errors.New("invalid token"))
			},
			expectedResult: nil,
			expectedError:  status.Error(codes.Unauthenticated, "недействительный токен доступа"),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var ctx context.Context
			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(context.Background(), tt.metadata)
			} else {
				ctx = context.Background()
			}

			unaryInterceptor := interceptor.Unary()

			result, err := unaryInterceptor(
				ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: tt.method},
				tt.handler,
			)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockValidator.AssertExpectations(t)
		})
	}
}
