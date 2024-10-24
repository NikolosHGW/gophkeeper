package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
)

type mockUpdateDataService struct {
	dataItem  *datapb.DataItem
	getErr    error
	updateErr error
}

func (m *mockUpdateDataService) GetData(ctx context.Context, token string, id int32) (*datapb.DataItem, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.dataItem, nil
}

func (m *mockUpdateDataService) UpdateData(ctx context.Context, token string, data *datapb.DataItem) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.dataItem = data
	return nil
}

func TestUpdateCommand_Execute_TokenMissing(t *testing.T) {
	dataService := &mockUpdateDataService{}
	tokenHolder := &entity.TokenHolder{Token: ""}
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err == nil || err.Error() != "вы должны войти в систему" {
		t.Errorf("Ожидалась ошибка 'вы должны войти в систему', получили: %v", err)
	}
}

func TestUpdateCommand_Execute_InvalidIDInput(t *testing.T) {
	dataService := &mockUpdateDataService{}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("abc\n")
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "некорректный ID") {
		t.Errorf("Ожидалась ошибка некорректного ID, получили: %v", err)
	}
}

func TestUpdateCommand_Execute_GetDataError(t *testing.T) {
	dataService := &mockUpdateDataService{
		getErr: errors.New("service error"),
	}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("1\n")
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "ошибка получения данных") {
		t.Errorf("Ожидалась ошибка получения данных, получили: %v", err)
	}
}

func TestUpdateCommand_Execute_UnknownInfoType(t *testing.T) {
	dataItem := &datapb.DataItem{
		Id:       1,
		InfoType: "unknown_type",
		Info:     []byte("{}"),
		Meta:     "meta data",
	}

	dataService := &mockUpdateDataService{
		dataItem: dataItem,
	}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := strings.NewReader("1\n")
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	output := writer.String()
	if !strings.Contains(output, "Неизвестный тип данных") {
		t.Errorf("Вывод не содержит 'Неизвестный тип данных', получили: %s", output)
	}
}

func TestUpdateCommand_Execute_UpdateLoginPasswordData(t *testing.T) {
	currentData := entity.LoginPasswordData{
		Login:    "current_user",
		Password: "current_pass",
		URL:      "http://current.com",
	}
	infoBytes, _ := json.Marshal(currentData)
	dataItem := &datapb.DataItem{
		Id:       1,
		InfoType: "login_password",
		Info:     infoBytes,
		Meta:     "current meta",
	}

	dataService := &mockUpdateDataService{
		dataItem: dataItem,
	}

	tokenHolder := &entity.TokenHolder{Token: "valid_token"}

	input := strings.Join([]string{
		"1",
		"new_user",
		"new_pass",
		"new_url",
		"new meta",
	}, "\n") + "\n"

	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	expectedUpdatedData := &entity.LoginPasswordData{
		Login:    "new_user",
		Password: "new_pass",
		URL:      "new_url",
	}
	expectedInfoBytes, _ := json.Marshal(expectedUpdatedData)

	if dataService.dataItem.InfoType != "login_password" {
		t.Errorf("Ожидался InfoType 'login_password', получили: %s", dataService.dataItem.InfoType)
	}

	if !bytes.Equal(dataService.dataItem.Info, expectedInfoBytes) {
		t.Errorf("Обновленные данные Info не совпадают с ожидаемыми")
	}

	if dataService.dataItem.Meta != "new meta" {
		t.Errorf("Ожидалась Meta 'new meta', получили: %s", dataService.dataItem.Meta)
	}

	output := writer.String()
	if !strings.Contains(output, "Данные успешно обновлены.") {
		t.Errorf("Вывод не содержит 'Данные успешно обновлены.', получили: %s", output)
	}
}

func TestUpdateCommand_Execute_UpdateTextData(t *testing.T) {
	currentData := entity.TextData{
		Text: "current text",
	}
	infoBytes, _ := json.Marshal(currentData)
	dataItem := &datapb.DataItem{
		Id:       2,
		InfoType: "text",
		Info:     infoBytes,
		Meta:     "current meta",
	}

	dataService := &mockUpdateDataService{
		dataItem: dataItem,
	}

	tokenHolder := &entity.TokenHolder{Token: "valid_token"}

	input := strings.Join([]string{
		"2",
		"new text",
		"new meta",
	}, "\n") + "\n"

	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	expectedUpdatedData := &entity.TextData{
		Text: "new text",
	}
	expectedInfoBytes, _ := json.Marshal(expectedUpdatedData)

	if dataService.dataItem.InfoType != "text" {
		t.Errorf("Ожидался InfoType 'text', получили: %s", dataService.dataItem.InfoType)
	}

	if !bytes.Equal(dataService.dataItem.Info, expectedInfoBytes) {
		t.Errorf("Обновленные данные Info не совпадают с ожидаемыми")
	}

	if dataService.dataItem.Meta != "new meta" {
		t.Errorf("Ожидалась Meta 'new meta', получили: %s", dataService.dataItem.Meta)
	}

	output := writer.String()
	if !strings.Contains(output, "Данные успешно обновлены.") {
		t.Errorf("Вывод не содержит 'Данные успешно обновлены.', получили: %s", output)
	}
}

func TestUpdateCommand_Execute_UpdateBinaryData(t *testing.T) {
	currentData := entity.BinaryData{
		FileName:    "current.bin",
		FileContent: []byte("current content"),
	}
	infoBytes, _ := json.Marshal(currentData)
	dataItem := &datapb.DataItem{
		Id:       3,
		InfoType: "binary",
		Info:     infoBytes,
		Meta:     "current meta",
	}

	dataService := &mockUpdateDataService{
		dataItem: dataItem,
	}

	tokenHolder := &entity.TokenHolder{Token: "valid_token"}

	newFileName := "newfile.bin"
	newFileContent := []byte("new content")
	os.WriteFile(newFileName, newFileContent, 0644)
	defer os.Remove(newFileName)

	input := strings.Join([]string{
		"3",
		newFileName,
		"new meta",
	}, "\n") + "\n"

	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	expectedUpdatedData := &entity.BinaryData{
		FileName:    newFileName,
		FileContent: newFileContent,
	}
	expectedInfoBytes, _ := json.Marshal(expectedUpdatedData)

	if dataService.dataItem.InfoType != "binary" {
		t.Errorf("Ожидался InfoType 'binary', получили: %s", dataService.dataItem.InfoType)
	}

	if !bytes.Equal(dataService.dataItem.Info, expectedInfoBytes) {
		t.Errorf("Обновленные данные Info не совпадают с ожидаемыми")
	}

	if dataService.dataItem.Meta != "new meta" {
		t.Errorf("Ожидалась Meta 'new meta', получили: %s", dataService.dataItem.Meta)
	}

	output := writer.String()
	if !strings.Contains(output, "Данные успешно обновлены.") {
		t.Errorf("Вывод не содержит 'Данные успешно обновлены.', получили: %s", output)
	}
}

func TestUpdateCommand_Execute_UpdateBankCardData(t *testing.T) {
	currentData := entity.BankCardData{
		CardNumber: "1234-5678-9012-3456",
		ExpiryDate: "12/24",
		CVV:        "123",
		HolderName: "Current Holder",
	}
	infoBytes, _ := json.Marshal(currentData)
	dataItem := &datapb.DataItem{
		Id:       4,
		InfoType: "bank_card",
		Info:     infoBytes,
		Meta:     "current meta",
	}

	dataService := &mockUpdateDataService{
		dataItem: dataItem,
	}

	tokenHolder := &entity.TokenHolder{Token: "valid_token"}

	input := strings.Join([]string{
		"4",
		"4321-8765-2109-6543",
		"11/25",
		"321",
		"New Holder",
		"new meta",
	}, "\n") + "\n"

	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err != nil {
		t.Fatalf("Не ожидалось ошибки, получили: %v", err)
	}

	expectedUpdatedData := &entity.BankCardData{
		CardNumber: "4321-8765-2109-6543",
		ExpiryDate: "11/25",
		CVV:        "321",
		HolderName: "New Holder",
	}
	expectedInfoBytes, _ := json.Marshal(expectedUpdatedData)

	if dataService.dataItem.InfoType != "bank_card" {
		t.Errorf("Ожидался InfoType 'bank_card', получили: %s", dataService.dataItem.InfoType)
	}

	if !bytes.Equal(dataService.dataItem.Info, expectedInfoBytes) {
		t.Errorf("Обновленные данные Info не совпадают с ожидаемыми")
	}

	if dataService.dataItem.Meta != "new meta" {
		t.Errorf("Ожидалась Meta 'new meta', получили: %s", dataService.dataItem.Meta)
	}

	output := writer.String()
	if !strings.Contains(output, "Данные успешно обновлены.") {
		t.Errorf("Вывод не содержит 'Данные успешно обновлены.', получили: %s", output)
	}
}

func TestUpdateCommand_Execute_ScannerError(t *testing.T) {
	dataService := &mockUpdateDataService{}
	tokenHolder := &entity.TokenHolder{Token: "valid_token"}
	reader := &errorReader{}
	writer := &bytes.Buffer{}

	updateCmd := NewUpdateCommand(dataService, tokenHolder, reader, writer)

	err := updateCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "ошибка ввода ID") {
		t.Errorf("Ожидалась ошибка ввода ID, получили: %v", err)
	}
}
