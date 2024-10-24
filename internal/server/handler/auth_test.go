package handler

import (
	"context"
	"errors"
	"testing"

	pb "github.com/NikolosHGW/goph-keeper/api/authpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Handle(ctx context.Context, req *pb.LoginUserRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func TestAuthServer_Login(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		name            string
		req             *pb.LoginUserRequest
		setupMock       func(m *MockAuthUseCase)
		expectedResp    *pb.LoginUserResponse
		expectedErrCode codes.Code
	}

	tests := []testCase{
		{
			name: "Успешная авторизация",
			req: &pb.LoginUserRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMock: func(m *MockAuthUseCase) {
				m.On("Handle", ctx, mock.AnythingOfType("*authpb.LoginUserRequest")).Return("testtoken", nil)
			},
			expectedResp: &pb.LoginUserResponse{
				BearerToken: "testtoken",
			},
			expectedErrCode: codes.OK,
		},
		{
			name: "Ошибка валидации - пустой логин",
			req: &pb.LoginUserRequest{
				Login:    "",
				Password: "password123",
			},
			setupMock:       func(m *MockAuthUseCase) {},
			expectedResp:    nil,
			expectedErrCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка валидации - пустой пароль",
			req: &pb.LoginUserRequest{
				Login:    "testuser",
				Password: "",
			},
			setupMock:       func(m *MockAuthUseCase) {},
			expectedResp:    nil,
			expectedErrCode: codes.InvalidArgument,
		},
		{
			name: "Слишком длинный пароль",
			req: &pb.LoginUserRequest{
				Login:    "testuser",
				Password: string(make([]byte, maxPasswordLength+1)),
			},
			setupMock:       func(m *MockAuthUseCase) {},
			expectedResp:    nil,
			expectedErrCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка в use case",
			req: &pb.LoginUserRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMock: func(m *MockAuthUseCase) {
				m.On("Handle", ctx, mock.AnythingOfType("*authpb.LoginUserRequest")).Return("", errors.New("some internal error"))
			},
			expectedResp:    nil,
			expectedErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthUseCase := new(MockAuthUseCase)
			if tt.setupMock != nil {
				tt.setupMock(mockAuthUseCase)
			}

			server := NewAuthServer(mockAuthUseCase)

			resp, err := server.LoginUser(ctx, tt.req)

			if tt.expectedErrCode != codes.OK {
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedErrCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}

			mockAuthUseCase.AssertExpectations(t)
		})
	}
}
