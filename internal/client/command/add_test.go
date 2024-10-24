package command

import (
	"bytes"
	"context"
	"encoding/json"
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
			name:  "Successful add - Login and Password",
			token: "valid_token",
			input: "1\nuser123\npass123\nhttp://example.com\nmeta_info\n",
			mockSetup: func(m *MockDataService) {
				loginPasswordData := &entity.LoginPasswordData{
					Login:    "user123",
					Password: "pass123",
					URL:      "http://example.com",
				}
				infoBytes, _ := json.Marshal(loginPasswordData)
				dataItem := &datapb.DataItem{
					InfoType: "login_password",
					Info:     infoBytes,
					Meta:     "meta_info",
				}
				m.On("AddData", mock.Anything, "valid_token", dataItem).Return(int32(1), nil)
			},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Введите логин: " +
				"Введите пароль: " +
				"Введите URL: " +
				"Введите метаинформацию: " +
				"Данные успешно добавлены с ID: 1\n",
			expectedError: nil,
		},
		{
			name:  "Successful add - Text",
			token: "valid_token",
			input: "2\nSample text data\nmeta_info_text\n",
			mockSetup: func(m *MockDataService) {
				textData := &entity.TextData{
					Text: "Sample text data",
				}
				infoBytes, _ := json.Marshal(textData)
				dataItem := &datapb.DataItem{
					InfoType: "text",
					Info:     infoBytes,
					Meta:     "meta_info_text",
				}
				m.On("AddData", mock.Anything, "valid_token", dataItem).Return(int32(2), nil)
			},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Введите текст: " +
				"Введите метаинформацию: " +
				"Данные успешно добавлены с ID: 2\n",
			expectedError: nil,
		},
		{
			name:  "Successful add - Bank Card",
			token: "valid_token",
			input: "4\n1234-5678-9012-3456\n12/24\n123\nJohn Doe\nmeta_info_card\n",
			mockSetup: func(m *MockDataService) {
				bankCardData := &entity.BankCardData{
					CardNumber: "1234-5678-9012-3456",
					ExpiryDate: "12/24",
					CVV:        "123",
					HolderName: "John Doe",
				}
				infoBytes, _ := json.Marshal(bankCardData)
				dataItem := &datapb.DataItem{
					InfoType: "bank_card",
					Info:     infoBytes,
					Meta:     "meta_info_card",
				}
				m.On("AddData", mock.Anything, "valid_token", dataItem).Return(int32(4), nil)
			},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Введите номер карты: " +
				"Введите срок действия (MM/YY): " +
				"Введите CVV: " +
				"Введите имя держателя карты: " +
				"Введите метаинформацию: " +
				"Данные успешно добавлены с ID: 4\n",
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
			name:      "Error reading option",
			token:     "valid_token",
			input:     "",
			mockSetup: func(m *MockDataService) {},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: ",
			expectedError: fmt.Errorf("ошибка ввода опции:"),
		},
		{
			name:      "Error reading login",
			token:     "valid_token",
			input:     "1\n",
			mockSetup: func(m *MockDataService) {},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Введите логин: ",
			expectedError: fmt.Errorf("ошибка ввода логина:"),
		},
		{
			name:  "Error adding data",
			token: "valid_token",
			input: "1\nuser123\npass123\nhttp://example.com\nmeta_info\n",
			mockSetup: func(m *MockDataService) {
				loginPasswordData := &entity.LoginPasswordData{
					Login:    "user123",
					Password: "pass123",
					URL:      "http://example.com",
				}
				infoBytes, _ := json.Marshal(loginPasswordData)
				dataItem := &datapb.DataItem{
					InfoType: "login_password",
					Info:     infoBytes,
					Meta:     "meta_info",
				}
				m.On("AddData", mock.Anything, "valid_token", dataItem).Return(int32(0), fmt.Errorf("service error"))
			},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Введите логин: " +
				"Введите пароль: " +
				"Введите URL: " +
				"Введите метаинформацию: ",
			expectedError: errors.New("ошибка добавления данных: service error"),
		},
		{
			name:      "Invalid option",
			token:     "valid_token",
			input:     "5\n",
			mockSetup: func(m *MockDataService) {},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Некорректная опция\n",
			expectedError: nil,
		},
		{
			name:      "Error reading text",
			token:     "valid_token",
			input:     "2\n",
			mockSetup: func(m *MockDataService) {},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Введите текст: ",
			expectedError: fmt.Errorf("ошибка ввода текста:"),
		},
		{
			name:      "Error reading file path",
			token:     "valid_token",
			input:     "3\n",
			mockSetup: func(m *MockDataService) {},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Введите путь к файлу: ",
			expectedError: fmt.Errorf("ошибка ввода пути к файлу:"),
		},
		{
			name:      "Error reading bank card number",
			token:     "valid_token",
			input:     "4\n",
			mockSetup: func(m *MockDataService) {},
			expectedOutput: "Выберите тип данных для добавления:\n" +
				"1. Login and Password\n" +
				"2. Text\n" +
				"3. Binary File\n" +
				"4. Bank Card\n" +
				"Введите номер опции: " +
				"Введите номер карты: ",
			expectedError: fmt.Errorf("ошибка ввода номера карты:"),
		},
	}

	for _, tt := range tests {
		tt := tt
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
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedOutput, writer.String())

			mockService.AssertExpectations(t)
		})
	}
}
