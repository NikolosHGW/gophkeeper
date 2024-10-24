package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockGetDataService struct {
	dataItem *datapb.DataItem
	err      error
}

func (m *mockGetDataService) GetData(ctx context.Context, token string, id int32) (*datapb.DataItem, error) {
	return m.dataItem, m.err
}

func TestGetCommand_Execute_TokenMissing(t *testing.T) {
	dataService := &mockGetDataService{}
	tokenHolder := &entity.TokenHolder{Token: ""}
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err == nil || err.Error() != "вы должны войти в систему" {
		t.Errorf("Ожидалась ошибка 'вы должны войти в систему', получили: %v", err)
	}
}

func TestGetCommand_Execute_InvalidIDInput(t *testing.T) {
	dataService := &mockGetDataService{}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("abc\n")
	writer := &bytes.Buffer{}

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "некорректный ID") {
		t.Errorf("Ожидалась ошибка некорректного ID, получили: %v", err)
	}
}

func TestGetCommand_Execute_GetDataError(t *testing.T) {
	dataService := &mockGetDataService{
		err: errors.New("service error"),
	}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("1\n")
	writer := &bytes.Buffer{}

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "ошибка получения данных") {
		t.Errorf("Ожидалась ошибка получения данных, получили: %v", err)
	}
}

func TestGetCommand_Execute_LoginPasswordData(t *testing.T) {
	loginPasswordData := entity.LoginPasswordData{
		Login:    "user",
		Password: "pass",
		URL:      "http://example.com",
	}
	infoBytes, _ := json.Marshal(loginPasswordData)

	dataItem := &datapb.DataItem{
		Id:       1,
		InfoType: "login_password",
		Info:     infoBytes,
		Meta:     "meta data",
		Created:  timestamppb.Now(),
	}

	dataService := &mockGetDataService{
		dataItem: dataItem,
	}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("1\n")
	writer := &bytes.Buffer{}

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	output := writer.String()
	if !strings.Contains(output, "Логин: user") {
		t.Errorf("Вывод не содержит ожидаемый логин, получили: %s", output)
	}
	if !strings.Contains(output, "Пароль: pass") {
		t.Errorf("Вывод не содержит ожидаемый пароль, получили: %s", output)
	}
	if !strings.Contains(output, "URL: http://example.com") {
		t.Errorf("Вывод не содержит ожидаемый URL, получили: %s", output)
	}
}

func TestGetCommand_Execute_TextData(t *testing.T) {
	textData := entity.TextData{
		Text: "some text",
	}
	infoBytes, _ := json.Marshal(textData)

	dataItem := &datapb.DataItem{
		Id:       2,
		InfoType: "text",
		Info:     infoBytes,
		Meta:     "meta data",
		Created:  timestamppb.Now(),
	}

	dataService := &mockGetDataService{
		dataItem: dataItem,
	}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("2\n")
	writer := &bytes.Buffer{}

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	output := writer.String()
	if !strings.Contains(output, "Текст: some text") {
		t.Errorf("Вывод не содержит ожидаемый текст, получили: %s", output)
	}
}

func TestGetCommand_Execute_BinaryData(t *testing.T) {
	binaryData := entity.BinaryData{
		FileName:    "testfile.bin",
		FileContent: []byte{0x00, 0x01, 0x02},
	}
	infoBytes, _ := json.Marshal(binaryData)

	dataItem := &datapb.DataItem{
		Id:       3,
		InfoType: "binary",
		Info:     infoBytes,
		Meta:     "meta data",
		Created:  timestamppb.Now(),
	}

	dataService := &mockGetDataService{
		dataItem: dataItem,
	}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("3\n")
	writer := &bytes.Buffer{}

	os.Remove("testfile.bin")

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	if _, err := os.Stat("testfile.bin"); os.IsNotExist(err) {
		t.Errorf("Файл testfile.bin не был создан")
	} else {
		os.Remove("testfile.bin")
	}

	output := writer.String()
	if !strings.Contains(output, "Бинарные данные сохранены в файл: testfile.bin") {
		t.Errorf("Вывод не содержит ожидаемое сообщение, получили: %s", output)
	}
}

func TestGetCommand_Execute_BankCardData(t *testing.T) {
	bankCardData := entity.BankCardData{
		CardNumber: "1234-5678-9012-3456",
		ExpiryDate: "12/24",
		CVV:        "123",
		HolderName: "John Doe",
	}
	infoBytes, _ := json.Marshal(bankCardData)

	dataItem := &datapb.DataItem{
		Id:       4,
		InfoType: "bank_card",
		Info:     infoBytes,
		Meta:     "meta data",
		Created:  timestamppb.Now(),
	}

	dataService := &mockGetDataService{
		dataItem: dataItem,
	}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("4\n")
	writer := &bytes.Buffer{}

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	output := writer.String()
	if !strings.Contains(output, "Номер карты: 1234-5678-9012-3456") {
		t.Errorf("Вывод не содержит ожидаемый номер карты, получили: %s", output)
	}
	if !strings.Contains(output, "Срок действия: 12/24") {
		t.Errorf("Вывод не содержит ожидаемый срок действия, получили: %s", output)
	}
	if !strings.Contains(output, "CVV: 123") {
		t.Errorf("Вывод не содержит ожидаемый CVV, получили: %s", output)
	}
	if !strings.Contains(output, "Имя держателя: John Doe") {
		t.Errorf("Вывод не содержит ожидаемое имя держателя, получили: %s", output)
	}
}

func TestGetCommand_Execute_UnknownInfoType(t *testing.T) {
	dataItem := &datapb.DataItem{
		Id:       5,
		InfoType: "unknown_type",
		Info:     []byte("{}"),
		Meta:     "meta data",
		Created:  timestamppb.Now(),
	}

	dataService := &mockGetDataService{
		dataItem: dataItem,
	}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("5\n")
	writer := &bytes.Buffer{}

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	output := writer.String()
	if !strings.Contains(output, "Неизвестный тип данных") {
		t.Errorf("Вывод не содержит сообщение о неизвестном типе данных, получили: %s", output)
	}
}

func TestGetCommand_Execute_ScannerError(t *testing.T) {
	dataService := &mockGetDataService{}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := &errorReader{}
	writer := &bytes.Buffer{}

	getCmd := NewGetCommand(dataService, tokenHolder, reader, writer)

	err := getCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "ошибка ввода ID") {
		t.Errorf("Ожидалась ошибка ввода ID, получили: %v", err)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func TestGetCommand_Name(t *testing.T) {
	cmd := NewGetCommand(nil, nil, nil, nil)
	expectedName := "get"
	actualName := cmd.Name()
	assert.Equal(t, expectedName, actualName, "Название команды должно быть 'get'")
}
