package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockGetDataService struct {
	mock.Mock
}

func (m *MockGetDataService) GetData(ctx context.Context, token string, id int32) (*datapb.DataItem, error) {
	args := m.Called(ctx, token, id)
	if dataItem, ok := args.Get(0).(*datapb.DataItem); ok {
		return dataItem, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestGetCommand_Execute(t *testing.T) {
	fixedTime := time.Date(2023, 10, 18, 12, 0, 0, 0, time.UTC)
	fixedTimestamp := &timestamppb.Timestamp{Seconds: fixedTime.Unix()}

	tests := []struct {
		name           string
		token          string
		input          string
		mockSetup      func(m *MockGetDataService)
		expectedOutput string
		expectedError  error
	}{
		{
			name:  "Успешное получение",
			token: "valid_token",
			input: "1\n",
			mockSetup: func(m *MockGetDataService) {
				dataItem := &datapb.DataItem{
					Id:       1,
					InfoType: "login_password",
					Info:     "user123",
					Meta:     "meta_info",
					Created:  fixedTimestamp,
				}
				m.On("GetData", mock.Anything, "valid_token", int32(1)).Return(dataItem, nil)
			},
			expectedOutput: "Введите ID данных: ID: 1\nТип: login_password\nДанные: user123\nМета: meta_info\nСоздано: 2023-10-18 12:00:00 +0000 UTC\n",
			expectedError:  nil,
		},
		{
			name:           "Без токена",
			token:          "",
			input:          "",
			mockSetup:      func(m *MockGetDataService) {},
			expectedOutput: "",
			expectedError:  errors.New("вы должны войти в систему"),
		},
		{
			name:           "Ошибка при чтении ID",
			token:          "valid_token",
			input:          "",
			mockSetup:      func(m *MockGetDataService) {},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("ошибка ввода ID"),
		},
		{
			name:           "Невалидный ID",
			token:          "valid_token",
			input:          "abc\n",
			mockSetup:      func(m *MockGetDataService) {},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("некорректный ID: strconv.ParseInt: parsing \"abc\": invalid syntax"),
		},
		{
			name:  "Ошибка при получении данных",
			token: "valid_token",
			input: "2\n",
			mockSetup: func(m *MockGetDataService) {
				m.On("GetData", mock.Anything, "valid_token", int32(2)).Return(nil, fmt.Errorf("service error"))
			},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("ошибка получения данных: service error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockGetDataService)
			tt.mockSetup(mockService)

			tokenHolder := &entity.TokenHolder{
				Token: tt.token,
			}

			reader := strings.NewReader(tt.input)
			var writer bytes.Buffer

			cmd := NewGetCommand(mockService, tokenHolder, reader, &writer)

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
