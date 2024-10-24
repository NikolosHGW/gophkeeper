package handler

import (
	"context"
	"errors"
	"testing"

	pb "github.com/NikolosHGW/goph-keeper/api/registerpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type registerUseCaseMock struct {
	handleFunc func(ctx context.Context, req *pb.RegisterUserRequest) (string, error)
}

func (m *registerUseCaseMock) Handle(ctx context.Context, req *pb.RegisterUserRequest) (string, error) {
	return m.handleFunc(ctx, req)
}

func TestRegisterServer_RegisterUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		req             *pb.RegisterUserRequest
		setupMock       func() *registerUseCaseMock
		expectedToken   string
		expectedErrCode codes.Code
	}{
		{
			name: "Успешная регистрация",
			req: &pb.RegisterUserRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMock: func() *registerUseCaseMock {
				return &registerUseCaseMock{
					handleFunc: func(ctx context.Context, req *pb.RegisterUserRequest) (string, error) {
						return "testtoken", nil
					},
				}
			},
			expectedToken:   "testtoken",
			expectedErrCode: codes.OK,
		},
		{
			name: "Ошибка валидации - пустой логин",
			req: &pb.RegisterUserRequest{
				Login:    "",
				Password: "password123",
			},
			setupMock:       func() *registerUseCaseMock { return nil },
			expectedErrCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка в use case",
			req: &pb.RegisterUserRequest{
				Login:    "testuser",
				Password: "password123",
			},
			setupMock: func() *registerUseCaseMock {
				return &registerUseCaseMock{
					handleFunc: func(ctx context.Context, req *pb.RegisterUserRequest) (string, error) {
						return "", errors.New("some error")
					},
				}
			},
			expectedErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *RegisterServer
			if tt.setupMock != nil {
				mockRegisterUseCase := tt.setupMock()
				server = NewRegisterServer(mockRegisterUseCase)
			} else {
				server = NewRegisterServer(nil)
			}

			resp, err := server.RegisterUser(ctx, tt.req)

			if tt.expectedErrCode != codes.OK {
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedErrCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, resp.BearerToken)
			}
		})
	}
}

func TestValidateRegisterUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *pb.RegisterUserRequest
		wantErr bool
	}{
		{
			name: "Валидный запрос",
			req: &pb.RegisterUserRequest{
				Login:    "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "Пустой логин",
			req: &pb.RegisterUserRequest{
				Login:    "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "Пустой пароль",
			req: &pb.RegisterUserRequest{
				Login:    "testuser",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "Слишком длинный пароль",
			req: &pb.RegisterUserRequest{
				Login:    "testuser",
				Password: string(make([]byte, maxPasswordLength+1)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoginPasswordRequest(tt.req.Login, tt.req.Password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRegisterUserRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
