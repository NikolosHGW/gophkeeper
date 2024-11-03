package usecase

import (
	"context"
	"errors"
	"testing"

	pb "github.com/NikolosHGW/goph-keeper/api/authpb"
	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *UserRepoMock) User(ctx context.Context, login string) (*entity.User, error) {
	args := m.Called(ctx, login)
	user, _ := args.Get(0).(*entity.User)
	return user, args.Error(1)
}

func TestAuth_Handle(t *testing.T) {
	mockRepo := new(UserRepoMock)
	mockTokenService := new(TokenServicerMock)

	authUseCase := NewAuth(mockTokenService, mockRepo)

	ctx := context.Background()
	req := &pb.LoginUserRequest{Login: "testuser"}

	type testCase struct {
		name             string
		setupMocks       func()
		expectedToken    string
		expectedError    error
		assertAdditional func()
	}

	tests := []testCase{
		{
			name: "успешная авторизация",
			setupMocks: func() {
				user := &entity.User{
					ID:       123,
					Login:    "testuser",
					Password: "hashedpassword",
				}
				token := "jwt.token.string"

				mockRepo.On("User", ctx, req.Login).Return(user, nil)
				mockTokenService.On("GenerateJWT", user).Return(token, nil)
			},
			expectedToken: "jwt.token.string",
			expectedError: nil,
			assertAdditional: func() {
				mockRepo.AssertExpectations(t)
				mockTokenService.AssertExpectations(t)
			},
		},
		{
			name: "пользователь не найден",
			setupMocks: func() {
				mockRepo.On("User", ctx, req.Login).Return(nil, helper.ErrInternalServer)
			},
			expectedToken: "",
			expectedError: helper.ErrInternalServer,
			assertAdditional: func() {
				mockRepo.AssertExpectations(t)
				mockTokenService.AssertNotCalled(t, "GenerateJWT", mock.Anything)
			},
		},
		{
			name: "ошибка при генерации токена",
			setupMocks: func() {
				user := &entity.User{
					ID:       123,
					Login:    "testuser",
					Password: "hashedpassword",
				}
				mockRepo.On("User", ctx, req.Login).Return(user, nil)
				mockTokenService.On("GenerateJWT", user).Return("", errors.New("генерация токена не удалась"))
			},
			expectedToken: "",
			expectedError: errors.New("ошибка при генерации токена: генерация токена не удалась"),
			assertAdditional: func() {
				mockRepo.AssertExpectations(t)
				mockTokenService.AssertExpectations(t)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			result, err := authUseCase.Handle(ctx, req)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
				assert.Equal(t, tc.expectedToken, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedToken, result)
			}

			if tc.assertAdditional != nil {
				tc.assertAdditional()
			}

			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			mockTokenService.ExpectedCalls = nil
			mockTokenService.Calls = nil
		})
	}
}
