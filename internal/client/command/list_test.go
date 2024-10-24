package command

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockListDataService struct {
	ListDataFunc func(ctx context.Context, token string, filter *entity.DataFilter) ([]*datapb.DataItem, error)
}

func (m *mockListDataService) ListData(ctx context.Context, token string, filter *entity.DataFilter) ([]*datapb.DataItem, error) {
	return m.ListDataFunc(ctx, token, filter)
}

func TestListCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		input          string
		listDataFunc   func(ctx context.Context, token string, filter *entity.DataFilter) ([]*datapb.DataItem, error)
		expectedOutput string
		expectedError  string
	}{
		{
			name:          "Пользователь не вошел в систему",
			token:         "",
			input:         "",
			expectedError: "вы должны войти в систему",
		},
		{
			name:          "Ошибка ввода",
			token:         "valid_token",
			input:         "",
			expectedError: "ошибка ввода типа данных:",
		},
		{
			name:  "Ошибка сервиса данных",
			token: "valid_token",
			input: "\n",
			listDataFunc: func(ctx context.Context, token string, filter *entity.DataFilter) ([]*datapb.DataItem, error) {
				return nil, errors.New("service error")
			},
			expectedError: "ошибка получения списка данных:",
		},
		{
			name:  "Пустой список данных",
			token: "valid_token",
			input: "\n",
			listDataFunc: func(ctx context.Context, token string, filter *entity.DataFilter) ([]*datapb.DataItem, error) {
				return []*datapb.DataItem{}, nil
			},
			expectedOutput: "Введите тип данных для фильтрации (оставьте пустым для всех типов): Данные не найдены.\n",
		},
		{
			name:  "Успешное получение данных",
			token: "valid_token",
			input: "\n",
			listDataFunc: func(ctx context.Context, token string, filter *entity.DataFilter) ([]*datapb.DataItem, error) {
				return []*datapb.DataItem{
					{
						Id:       1,
						InfoType: "type1",
						Meta:     "meta1",
						Created:  timestamppb.New(time.Date(2023, 10, 15, 12, 0, 0, 0, time.UTC)),
					},
					{
						Id:       2,
						InfoType: "type2",
						Meta:     "meta2",
						Created:  timestamppb.New(time.Date(2023, 10, 16, 13, 0, 0, 0, time.UTC)),
					},
				}, nil
			},
			expectedOutput: "Введите тип данных для фильтрации (оставьте пустым для всех типов): Список данных:\n" +
				"ID: 1, Тип: type1, Мета: meta1, Дата создания: 2023-10-15 12:00:00\n" +
				"ID: 2, Тип: type2, Мета: meta2, Дата создания: 2023-10-16 13:00:00\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			writer := &bytes.Buffer{}

			tokenHolder := &entity.TokenHolder{Token: tt.token}

			dataService := &mockListDataService{
				ListDataFunc: tt.listDataFunc,
			}

			if tt.listDataFunc == nil {
				dataService.ListDataFunc = func(ctx context.Context, token string, filter *entity.DataFilter) ([]*datapb.DataItem, error) {
					return nil, nil
				}
			}

			cmd := NewListCommand(dataService, tokenHolder, reader, writer)

			err := cmd.Execute()

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
				return
			}

			output := writer.String()
			if output != tt.expectedOutput {
				t.Errorf("ожидался вывод:\n%s\nполучено:\n%s", tt.expectedOutput, output)
			}
		})
	}
}

func TestListCommand_Name(t *testing.T) {
	cmd := NewListCommand(nil, nil, nil, nil)
	expectedName := "list"
	actualName := cmd.Name()
	assert.Equal(t, expectedName, actualName, "Название команды должно быть 'list'")
}
