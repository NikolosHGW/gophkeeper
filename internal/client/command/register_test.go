package command

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

func TestRegisterCommand_Execute_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	expectedToken := "mocked_token"
	mockAuthService.On("Register", mock.Anything, "testuser", "testpass").Return(expectedToken, nil)

	tokenHolder := &entity.TokenHolder{}

	input := "testuser\ntestpass\n"
	reader := bytes.NewBufferString(input)
	writer := &bytes.Buffer{}

	cmd := NewRegisterCommand(mockAuthService, tokenHolder, reader, writer)

	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, tokenHolder.Token)
	assert.Contains(t, writer.String(), "Регистрация прошла успешно.")

	mockAuthService.AssertExpectations(t)
}

func TestRegisterCommand_Execute_RegisterError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockAuthService.On("Register", mock.Anything, "testuser", "wrongpass").Return("", errors.New("registration failed"))

	tokenHolder := &entity.TokenHolder{}

	input := "testuser\nwrongpass\n"
	reader := bytes.NewBufferString(input)
	writer := &bytes.Buffer{}

	cmd := NewRegisterCommand(mockAuthService, tokenHolder, reader, writer)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка регистрации")
	assert.Empty(t, tokenHolder.Token)

	assert.NotContains(t, writer.String(), "Регистрация прошла успешно.")

	mockAuthService.AssertExpectations(t)
}

func TestRegisterCommand_Execute_InputError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	tokenHolder := &entity.TokenHolder{}

	reader := bytes.NewBuffer(nil)
	writer := &bytes.Buffer{}

	cmd := NewRegisterCommand(mockAuthService, tokenHolder, reader, writer)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка ввода логина")
	assert.Empty(t, tokenHolder.Token)
}
