package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockUpdateDataService struct {
	mock.Mock
}

func (m *MockUpdateDataService) GetData(ctx context.Context, token string, id int32) (*datapb.DataItem, error) {
	args := m.Called(ctx, token, id)
	if dataItem, ok := args.Get(0).(*datapb.DataItem); ok {
		return dataItem, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUpdateDataService) UpdateData(ctx context.Context, token string, data *datapb.DataItem) error {
	args := m.Called(ctx, token, data)
	return args.Error(0)
}

type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

func TestUpdateCommand_Execute(t *testing.T) {
	fixedTime := time.Date(2023, 10, 18, 12, 0, 0, 0, time.UTC)
	fixedTimestamp := &timestamppb.Timestamp{Seconds: fixedTime.Unix()}

	tests := []struct {
		name           string
		token          string
		input          string
		mockSetup      func(m *MockUpdateDataService)
		expectedOutput string
		expectedError  error
	}{
		{
			name:  "Успешное обновление всех полей",
			token: "valid_token",
			input: "1\nnew_type\nnew_info\nnew_meta\n",
			mockSetup: func(m *MockUpdateDataService) {
				originalData := &datapb.DataItem{
					Id:       1,
					InfoType: "login_password",
					Info:     "user123",
					Meta:     "meta_info",
					Created:  fixedTimestamp,
				}
				m.On("GetData", mock.Anything, "valid_token", int32(1)).Return(originalData, nil)

				updatedData := &datapb.DataItem{
					Id:       1,
					InfoType: "new_type",
					Info:     "new_info",
					Meta:     "new_meta",
					Created:  fixedTimestamp,
				}
				m.On("UpdateData", mock.Anything, "valid_token", updatedData).Return(nil)
			},
			expectedOutput: "Введите ID данных: Текущий тип (login_password): Текущие данные (user123): Текущая мета (meta_info): Данные успешно обновлены.\n",
			expectedError:  nil,
		},
		{
			name:           "Без токена",
			token:          "",
			input:          "",
			mockSetup:      func(m *MockUpdateDataService) {},
			expectedOutput: "",
			expectedError:  errors.New("вы должны войти в систему"),
		},
		{
			name:           "Ошибка при записи ID",
			token:          "valid_token",
			input:          "",
			mockSetup:      func(m *MockUpdateDataService) {},
			expectedOutput: "",
			expectedError:  errors.New("ошибка stdin ID: write error"),
		},
		{
			name:           "Ошибка при чтении ID",
			token:          "valid_token",
			input:          "",
			mockSetup:      func(m *MockUpdateDataService) {},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("ошибка ввода ID"),
		},
		{
			name:           "Невалидный токен",
			token:          "valid_token",
			input:          "abc\n",
			mockSetup:      func(m *MockUpdateDataService) {},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("некорректный ID: strconv.ParseInt: parsing \"abc\": invalid syntax"),
		},
		{
			name:  "Ошибка при получении данных",
			token: "valid_token",
			input: "2\n",
			mockSetup: func(m *MockUpdateDataService) {
				m.On("GetData", mock.Anything, "valid_token", int32(2)).Return(nil, fmt.Errorf("service error"))
			},
			expectedOutput: "Введите ID данных: ",
			expectedError:  errors.New("ошибка получения данных: service error"),
		},
		{
			name:  "Ошибка при обновлении данных",
			token: "valid_token",
			input: "3\nnew_type\nnew_info\nnew_meta\n",
			mockSetup: func(m *MockUpdateDataService) {
				originalData := &datapb.DataItem{
					Id:       3,
					InfoType: "original_type",
					Info:     "original_info",
					Meta:     "original_meta",
					Created:  fixedTimestamp,
				}
				m.On("GetData", mock.Anything, "valid_token", int32(3)).Return(originalData, nil)

				updatedData := &datapb.DataItem{
					Id:       3,
					InfoType: "new_type",
					Info:     "new_info",
					Meta:     "new_meta",
					Created:  fixedTimestamp,
				}
				m.On("UpdateData", mock.Anything, "valid_token", updatedData).Return(fmt.Errorf("update error"))
			},
			expectedOutput: "Введите ID данных: Текущий тип (original_type): Текущие данные (original_info): Текущая мета (original_meta): ",
			expectedError:  errors.New("ошибка обновления данных: update error"),
		},
		{
			name:  "Пустые поля сохраняют исходные значения",
			token: "valid_token",
			input: "4\n\n\n\n",
			mockSetup: func(m *MockUpdateDataService) {
				originalData := &datapb.DataItem{
					Id:       4,
					InfoType: "original_type",
					Info:     "original_info",
					Meta:     "original_meta",
					Created:  fixedTimestamp,
				}
				m.On("GetData", mock.Anything, "valid_token", int32(4)).Return(originalData, nil)

				updatedData := &datapb.DataItem{
					Id:       4,
					InfoType: "original_type",
					Info:     "original_info",
					Meta:     "original_meta",
					Created:  fixedTimestamp,
				}
				m.On("UpdateData", mock.Anything, "valid_token", updatedData).Return(nil)
			},
			expectedOutput: "Введите ID данных: Текущий тип (original_type): Текущие данные (original_info): Текущая мета (original_meta): Данные успешно обновлены.\n",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUpdateDataService)
			tt.mockSetup(mockService)

			tokenHolder := &entity.TokenHolder{
				Token: tt.token,
			}

			var writer io.Writer
			var outputBuffer bytes.Buffer

			if tt.name == "Ошибка при записи ID" {
				writer = &errorWriter{err: errors.New("write error")}
			} else {
				writer = &outputBuffer
			}

			reader := strings.NewReader(tt.input)

			cmd := NewUpdateCommand(mockService, tokenHolder, reader, writer)

			err := cmd.Execute()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.name == "Error writing ID prompt" {
				assert.Equal(t, "", outputBuffer.String())
			} else {
				assert.Equal(t, tt.expectedOutput, outputBuffer.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
