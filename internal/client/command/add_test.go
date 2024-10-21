package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDataService struct {
	mock.Mock
}

func (m *MockDataService) AddData(ctx context.Context, token string, data *datapb.DataItem) (int32, error) {
	args := m.Called(ctx, token, data)
	return args.Get(0).(int32), args.Error(1)
}

func TestAddCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		input          string
		mockSetup      func(m *MockDataService)
		expectedOutput string
		expectedError  error
	}{
		{
			name:  "Successful add",
			token: "valid_token",
			input: "login_password\nuser123\nmeta_info\n",
			mockSetup: func(m *MockDataService) {
				dataItem := &datapb.DataItem{
					InfoType: "login_password",
					Info:     "user123",
					Meta:     "meta_info",
				}
				m.On("AddData", mock.Anything, "valid_token", dataItem).Return(int32(1), nil)
			},
			expectedOutput: "Введите тип информации (login_password, text, binary, bank_card): " +
				"Введите данные: " +
				"Введите метаинформацию: " +
				"Данные успешно добавлены с ID: 1\n",
			expectedError: nil,
		},
		{
			name:           "Missing token",
			token:          "",
			input:          "",
			mockSetup:      func(m *MockDataService) {},
			expectedOutput: "",
			expectedError:  errors.New("вы должны войти в систему"),
		},
		{
			name:           "Error reading info type",
			token:          "valid_token",
			input:          "",
			mockSetup:      func(m *MockDataService) {},
			expectedOutput: "Введите тип информации (login_password, text, binary, bank_card): ",
			expectedError:  errors.New("ошибка ввода типа информации: unexpected EOF"),
		},
		{
			name:      "Error reading data",
			token:     "valid_token",
			input:     "login_password\n",
			mockSetup: func(m *MockDataService) {},
			expectedOutput: "Введите тип информации (login_password, text, binary, bank_card): " +
				"Введите данные: ",
			expectedError: errors.New("ошибка ввода пароля: unexpected EOF"),
		},
		{
			name:      "Error reading meta information",
			token:     "valid_token",
			input:     "login_password\nuser123\n",
			mockSetup: func(m *MockDataService) {},
			expectedOutput: "Введите тип информации (login_password, text, binary, bank_card): " +
				"Введите данные: " +
				"Введите метаинформацию: ",
			expectedError: errors.New("ошибка ввода метаинформации: unexpected EOF"),
		},
		{
			name:  "Error adding data",
			token: "valid_token",
			input: "login_password\nuser123\nmeta_info\n",
			mockSetup: func(m *MockDataService) {
				dataItem := &datapb.DataItem{
					InfoType: "login_password",
					Info:     "user123",
					Meta:     "meta_info",
				}
				m.On("AddData", mock.Anything, "valid_token", dataItem).Return(int32(0), fmt.Errorf("service error"))
			},
			expectedOutput: "Введите тип информации (login_password, text, binary, bank_card): " +
				"Введите данные: " +
				"Введите метаинформацию: ",
			expectedError: errors.New("ошибка добавления данных: service error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockDataService)
			tt.mockSetup(mockService)

			tokenHolder := &entity.TokenHolder{
				Token: tt.token,
			}

			reader := strings.NewReader(tt.input)
			var writer bytes.Buffer

			cmd := NewAddCommand(mockService, tokenHolder, reader, &writer)

			err := cmd.Execute()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedOutput, writer.String())

			mockService.AssertExpectations(t)
		})
	}
}
