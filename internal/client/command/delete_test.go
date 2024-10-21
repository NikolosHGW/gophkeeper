package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDeleteDataService struct {
	mock.Mock
}

func (m *MockDeleteDataService) DeleteData(ctx context.Context, token string, id int32) error {
	args := m.Called(ctx, token, id)
	return args.Error(0)
}

func TestDeleteCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		input          string
		mockSetup      func(m *MockDeleteDataService)
		expectedOutput string
		expectedError  error
	}{
		{
			name:  "Успешное удаление",
			token: "valid_token",
			input: "1\n",
			mockSetup: func(m *MockDeleteDataService) {
				m.On("DeleteData", context.Background(), "valid_token", int32(1)).Return(nil)
			},
			expectedOutput: "Введите ID данных: Данные успешно удалены.\n",
			expectedError:  nil,
		},
		{
			name:           "Отсутствие токена",
			token:          "",
			input:          "",
			mockSetup:      func(m *MockDeleteDataService) {},
			expectedOutput: "",
			expectedError:  errors.New("вы должны войти в систему"),
		},
		{
			name:           "Ошибка вывода запроса на ввод ID",
			token:          "valid_token",
			input:          "",
			mockSetup:      func(m *MockDeleteDataService) {},
			expectedOutput: "",
			expectedError:  errors.New("ошибка вывода запроса на ввод ID: write error"),
		},
		{
			name:           "Ошибка ввода ID",
			token:          "valid_token",
			input:          "",
			mockSetup:      func(m *MockDeleteDataService) {},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("ошибка ввода ID"),
		},
		{
			name:           "Некорректный формат ID",
			token:          "valid_token",
			input:          "abc\n",
			mockSetup:      func(m *MockDeleteDataService) {},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("некорректный ID: strconv.ParseInt: parsing \"abc\": invalid syntax"),
		},
		{
			name:  "Ошибка удаления данных",
			token: "valid_token",
			input: "2\n",
			mockSetup: func(m *MockDeleteDataService) {
				m.On("DeleteData", context.Background(), "valid_token", int32(2)).Return(fmt.Errorf("service error"))
			},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("ошибка удаления данных: service error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockDeleteDataService)
			tt.mockSetup(mockService)

			tokenHolder := &entity.TokenHolder{
				Token: tt.token,
			}

			reader := strings.NewReader(tt.input)
			var writer io.Writer
			var outputBuffer bytes.Buffer

			if tt.name == "Ошибка вывода запроса на ввод ID" {
				writer = &errorWriter{err: errors.New("write error")}
			} else {
				writer = &outputBuffer
			}

			cmd := NewDeleteCommand(mockService, tokenHolder, reader, writer)

			err := cmd.Execute()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.name == "Ошибка вывода запроса на ввод ID" {
				assert.Equal(t, "", outputBuffer.String())
			} else {
				assert.Equal(t, tt.expectedOutput, outputBuffer.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
