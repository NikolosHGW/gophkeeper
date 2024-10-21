package usecase

import (
	"context"
	"errors"
	"testing"

	pb "github.com/NikolosHGW/goph-keeper/api/registerpb"
	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UserRepoMock struct {
	mock.Mock
}

func (m *UserRepoMock) Save(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepoMock) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	args := m.Called(ctx, login)
	return args.Bool(0), args.Error(1)
}

type RegisterServicerMock struct {
	mock.Mock
}

func (m *RegisterServicerMock) CreateUser(req *pb.RegisterUserRequest) (*entity.User, error) {
	args := m.Called(req)
	return args.Get(0).(*entity.User), args.Error(1)
}

type TokenServicerMock struct {
	mock.Mock
}

func (m *TokenServicerMock) GenerateJWT(user *entity.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func TestRegister_Handle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		setupMocks    func(*UserRepoMock, *RegisterServicerMock, *TokenServicerMock)
		req           *pb.RegisterUserRequest
		expectedToken string
		expectedError error
	}{
		{
			name: "Successful registration",
			setupMocks: func(userRepo *UserRepoMock, registerService *RegisterServicerMock, tokenService *TokenServicerMock) {
				userRepo.On("ExistsByLogin", ctx, "newuser").Return(false, nil)
				user := &entity.User{Login: "newuser", Password: "hashedpassword"}
				registerService.On("CreateUser", mock.Anything).Return(user, nil)
				userRepo.On("Save", ctx, user).Return(nil)
				tokenService.On("GenerateJWT", user).Return("token123", nil)
			},
			req: &pb.RegisterUserRequest{
				Login:    "newuser",
				Password: "password123",
			},
			expectedToken: "token123",
			expectedError: nil,
		},
		{
			name: "Login already exists",
			setupMocks: func(userRepo *UserRepoMock, registerService *RegisterServicerMock, tokenService *TokenServicerMock) {
				userRepo.On("ExistsByLogin", ctx, "existinguser").Return(true, nil)
			},
			req: &pb.RegisterUserRequest{
				Login:    "existinguser",
				Password: "password123",
			},
			expectedToken: "",
			expectedError: helper.ErrLoginAlreadyExists,
		},
		{
			name: "Error checking login existence",
			setupMocks: func(userRepo *UserRepoMock, registerService *RegisterServicerMock, tokenService *TokenServicerMock) {
				userRepo.On("ExistsByLogin", ctx, "newuser").Return(false, errors.New("database error"))
			},
			req: &pb.RegisterUserRequest{
				Login:    "newuser",
				Password: "password123",
			},
			expectedToken: "",
			expectedError: helper.ErrInternalServer,
		},
		{
			name: "Error creating user",
			setupMocks: func(userRepo *UserRepoMock, registerService *RegisterServicerMock, tokenService *TokenServicerMock) {
				userRepo.On("ExistsByLogin", ctx, "newuser").Return(false, nil)
				registerService.On("CreateUser", mock.Anything).Return((*entity.User)(nil), errors.New("creation error"))
			},
			req: &pb.RegisterUserRequest{
				Login:    "newuser",
				Password: "password123",
			},
			expectedToken: "",
			expectedError: errors.New("ошибка создания пользователя: creation error"),
		},
		{
			name: "Error saving user",
			setupMocks: func(userRepo *UserRepoMock, registerService *RegisterServicerMock, tokenService *TokenServicerMock) {
				userRepo.On("ExistsByLogin", ctx, "newuser").Return(false, nil)
				user := &entity.User{Login: "newuser", Password: "hashedpassword"}
				registerService.On("CreateUser", mock.Anything).Return(user, nil)
				userRepo.On("Save", ctx, user).Return(errors.New("save error"))
			},
			req: &pb.RegisterUserRequest{
				Login:    "newuser",
				Password: "password123",
			},
			expectedToken: "",
			expectedError: errors.New("ошибка при сохранении пользователя: save error"),
		},
		{
			name: "Error generating token",
			setupMocks: func(userRepo *UserRepoMock, registerService *RegisterServicerMock, tokenService *TokenServicerMock) {
				userRepo.On("ExistsByLogin", ctx, "newuser").Return(false, nil)
				user := &entity.User{Login: "newuser", Password: "hashedpassword"}
				registerService.On("CreateUser", mock.Anything).Return(user, nil)
				userRepo.On("Save", ctx, user).Return(nil)
				tokenService.On("GenerateJWT", user).Return("", errors.New("token error"))
			},
			req: &pb.RegisterUserRequest{
				Login:    "newuser",
				Password: "password123",
			},
			expectedToken: "",
			expectedError: errors.New("ошибка при генерации токена: token error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepoMock := new(UserRepoMock)
			registerServiceMock := new(RegisterServicerMock)
			tokenServiceMock := new(TokenServicerMock)

			tt.setupMocks(userRepoMock, registerServiceMock, tokenServiceMock)

			reg := NewRegister(registerServiceMock, tokenServiceMock, userRepoMock)

			token, err := reg.Handle(ctx, tt.req)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}

			userRepoMock.AssertExpectations(t)
			registerServiceMock.AssertExpectations(t)
			tokenServiceMock.AssertExpectations(t)
		})
	}
}
