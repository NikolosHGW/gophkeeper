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

type MockService struct {
	mock.Mock
}

func (m *MockService) Login(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

func TestLoginCommand_Execute_Success(t *testing.T) {
	mockService := new(MockService)
	expectedToken := "mocked_token"
	mockService.On("Login", mock.Anything, "testuser", "testpass").Return(expectedToken, nil)

	tokenHolder := &entity.TokenHolder{}

	input := "testuser\ntestpass\n"
	reader := bytes.NewBufferString(input)
	writer := &bytes.Buffer{}

	cmd := NewLoginCommand(mockService, tokenHolder, reader, writer)

	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, tokenHolder.Token)
}

func TestLoginCommand_Execute_AuthError(t *testing.T) {
	mockService := new(MockService)
	mockService.On("Login", mock.Anything, "testuser", "wrongpass").Return("", errors.New("authentication failed"))

	tokenHolder := &entity.TokenHolder{}

	input := "testuser\nwrongpass\n"
	reader := bytes.NewBufferString(input)
	writer := &bytes.Buffer{}

	cmd := NewLoginCommand(mockService, tokenHolder, reader, writer)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка входа")
	assert.Empty(t, tokenHolder.Token)
}

func TestLoginCommand_Execute_InputError(t *testing.T) {
	mockService := new(MockService)
	tokenHolder := &entity.TokenHolder{}

	t.Run("Ошибка ввода логина", func(t *testing.T) {
		reader := bytes.NewBuffer(nil)
		writer := &bytes.Buffer{}

		cmd := NewLoginCommand(mockService, tokenHolder, reader, writer)

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка ввода логина")
		assert.Empty(t, tokenHolder.Token)
	})

	t.Run("Ошибка ввода пароля", func(t *testing.T) {
		input := "testuser\n"
		reader := bytes.NewBufferString(input)
		writer := &bytes.Buffer{}

		cmd := NewLoginCommand(mockService, tokenHolder, reader, writer)

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка ввода пароля")
		assert.Empty(t, tokenHolder.Token)
	})
}

func TestLoginCommand_Execute_WriteError(t *testing.T) {
	mockService := new(MockService)
	tokenHolder := &entity.TokenHolder{}

	errorWriter := &ErrorWriter{}

	input := "testuser\ntestpass\n"
	reader := bytes.NewBufferString(input)
	writer := errorWriter

	cmd := NewLoginCommand(mockService, tokenHolder, reader, writer)

	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка stdin login")
	assert.Empty(t, tokenHolder.Token)
}

// ErrorWriter — это io.Writer, который всегда возвращает ошибку.
type ErrorWriter struct{}

func (e *ErrorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}
