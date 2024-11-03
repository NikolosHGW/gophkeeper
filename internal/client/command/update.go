package command

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
)

type updateDataService interface {
	GetData(ctx context.Context, token string, id int32) (*datapb.DataItem, error)
	UpdateData(ctx context.Context, token string, data *datapb.DataItem) error
}

type UpdateCommand struct {
	dataService updateDataService
	tokenHolder *entity.TokenHolder
	reader      io.Reader
	writer      io.Writer
}

func NewUpdateCommand(
	dataService updateDataService,
	tokenHolder *entity.TokenHolder,
	reader io.Reader,
	writer io.Writer,
) *UpdateCommand {
	return &UpdateCommand{
		dataService: dataService,
		tokenHolder: tokenHolder,
		reader:      reader,
		writer:      writer,
	}
}

func (c *UpdateCommand) Name() string {
	return "update"
}

func (c *UpdateCommand) Execute() error {
	if c.tokenHolder.Token == "" {
		return fmt.Errorf("вы должны войти в систему")
	}

	_, err := fmt.Fprint(c.writer, "Введите ID данных: ")
	if err != nil {
		return fmt.Errorf("ошибка stdin ID: %w", err)
	}

	scanner := bufio.NewScanner(c.reader)
	if !scanner.Scan() {
		return fmt.Errorf("ошибка ввода ID")
	}

	idStr := scanner.Text()
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return fmt.Errorf("некорректный ID: %w", err)
	}
	id := int32(id64)

	dataItem, err := c.dataService.GetData(context.Background(), c.tokenHolder.Token, id)
	if err != nil {
		return fmt.Errorf("ошибка получения данных: %w", err)
	}

	var updatedDataItem *datapb.DataItem

	switch dataItem.InfoType {
	case "login_password":
		updatedDataItem, err = c.updateLoginPasswordData(scanner, dataItem)
	case "text":
		updatedDataItem, err = c.updateTextData(scanner, dataItem)
	case "binary":
		updatedDataItem, err = c.updateBinaryData(scanner, dataItem)
	case "bank_card":
		updatedDataItem, err = c.updateBankCardData(scanner, dataItem)
	default:
		fmt.Fprintln(c.writer, "Неизвестный тип данных")
		return nil
	}

	if err != nil {
		return err
	}

	err = c.dataService.UpdateData(context.Background(), c.tokenHolder.Token, updatedDataItem)
	if err != nil {
		return fmt.Errorf("ошибка обновления данных: %w", err)
	}

	fmt.Fprintln(c.writer, "Данные успешно обновлены.")
	return nil
}

func (c *UpdateCommand) updateLoginPasswordData(scanner *bufio.Scanner, dataItem *datapb.DataItem) (*datapb.DataItem, error) {
	var currentData entity.LoginPasswordData
	if err := json.Unmarshal(dataItem.Info, &currentData); err != nil {
		return nil, fmt.Errorf("ошибка десериализации текущих данных: %w", err)
	}

	fmt.Fprintf(c.writer, "Текущий логин: %s\n", currentData.Login)
	fmt.Fprint(c.writer, "Введите новый логин (оставьте пустым, чтобы оставить без изменений): ")
	var login string
	if scanner.Scan() {
		login = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода логина: %w", scanner.Err())
	}
	if login == "" {
		login = currentData.Login
	}

	fmt.Fprintf(c.writer, "Текущий пароль: %s\n", currentData.Password)
	fmt.Fprint(c.writer, "Введите новый пароль (оставьте пустым, чтобы оставить без изменений): ")
	var password string
	if scanner.Scan() {
		password = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода пароля: %w", scanner.Err())
	}
	if password == "" {
		password = currentData.Password
	}

	fmt.Fprintf(c.writer, "Текущий URL: %s\n", currentData.URL)
	fmt.Fprint(c.writer, "Введите новый URL (оставьте пустым, чтобы оставить без изменений): ")
	var url string
	if scanner.Scan() {
		url = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода URL: %w", scanner.Err())
	}
	if url == "" {
		url = currentData.URL
	}

	fmt.Fprintf(c.writer, "Текущая метаинформация: %s\n", dataItem.Meta)
	fmt.Fprint(c.writer, "Введите новую метаинформацию (оставьте пустым, чтобы оставить без изменений): ")
	var meta string
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода метаинформации: %w", scanner.Err())
	}
	if meta == "" {
		meta = dataItem.Meta
	}

	updatedData := &entity.LoginPasswordData{
		Login:    login,
		Password: password,
		URL:      url,
	}

	infoBytes, err := json.Marshal(updatedData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации обновленных данных: %w", err)
	}

	updatedDataItem := &datapb.DataItem{
		Id:       dataItem.Id,
		InfoType: dataItem.InfoType,
		Info:     infoBytes,
		Meta:     meta,
	}

	return updatedDataItem, nil
}

func (c *UpdateCommand) updateTextData(scanner *bufio.Scanner, dataItem *datapb.DataItem) (*datapb.DataItem, error) {
	var currentData entity.TextData
	if err := json.Unmarshal(dataItem.Info, &currentData); err != nil {
		return nil, fmt.Errorf("ошибка десериализации текущих данных: %w", err)
	}

	fmt.Fprintf(c.writer, "Текущий текст: %s\n", currentData.Text)
	fmt.Fprint(c.writer, "Введите новый текст (оставьте пустым, чтобы оставить без изменений): ")
	var text string
	if scanner.Scan() {
		text = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода текста: %w", scanner.Err())
	}
	if text == "" {
		text = currentData.Text
	}

	fmt.Fprintf(c.writer, "Текущая метаинформация: %s\n", dataItem.Meta)
	fmt.Fprint(c.writer, "Введите новую метаинформацию (оставьте пустым, чтобы оставить без изменений): ")
	var meta string
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода метаинформации: %w", scanner.Err())
	}
	if meta == "" {
		meta = dataItem.Meta
	}

	updatedData := &entity.TextData{
		Text: text,
	}

	infoBytes, err := json.Marshal(updatedData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации обновленных данных: %w", err)
	}

	updatedDataItem := &datapb.DataItem{
		Id:       dataItem.Id,
		InfoType: dataItem.InfoType,
		Info:     infoBytes,
		Meta:     meta,
	}

	return updatedDataItem, nil
}

func (c *UpdateCommand) updateBinaryData(scanner *bufio.Scanner, dataItem *datapb.DataItem) (*datapb.DataItem, error) {
	var currentData entity.BinaryData
	if err := json.Unmarshal(dataItem.Info, &currentData); err != nil {
		return nil, fmt.Errorf("ошибка десериализации текущих данных: %w", err)
	}

	fmt.Fprintf(c.writer, "Текущее имя файла: %s\n", currentData.FileName)
	fmt.Fprint(c.writer, "Введите новый путь к файлу для обновления (оставьте пустым, чтобы оставить без изменений): ")
	var filePath string
	if scanner.Scan() {
		filePath = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода пути к файлу: %w", scanner.Err())
	}

	var fileName string
	var fileContent []byte
	if filePath != "" {
		fileContent_, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
		}
		fileName = getFileName(filePath)
		fileContent = fileContent_
	} else {
		fileName = currentData.FileName
		fileContent = currentData.FileContent
	}

	fmt.Fprintf(c.writer, "Текущая метаинформация: %s\n", dataItem.Meta)
	fmt.Fprint(c.writer, "Введите новую метаинформацию (оставьте пустым, чтобы оставить без изменений): ")
	var meta string
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода метаинформации: %w", scanner.Err())
	}
	if meta == "" {
		meta = dataItem.Meta
	}

	updatedData := &entity.BinaryData{
		FileName:    fileName,
		FileContent: fileContent,
	}

	infoBytes, err := json.Marshal(updatedData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации обновленных данных: %w", err)
	}

	updatedDataItem := &datapb.DataItem{
		Id:       dataItem.Id,
		InfoType: dataItem.InfoType,
		Info:     infoBytes,
		Meta:     meta,
	}

	return updatedDataItem, nil
}

func (c *UpdateCommand) updateBankCardData(scanner *bufio.Scanner, dataItem *datapb.DataItem) (*datapb.DataItem, error) {
	var currentData entity.BankCardData
	if err := json.Unmarshal(dataItem.Info, &currentData); err != nil {
		return nil, fmt.Errorf("ошибка десериализации текущих данных: %w", err)
	}

	fmt.Fprintf(c.writer, "Текущий номер карты: %s\n", currentData.CardNumber)
	fmt.Fprint(c.writer, "Введите новый номер карты (оставьте пустым, чтобы оставить без изменений): ")
	var cardNumber string
	if scanner.Scan() {
		cardNumber = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода номера карты: %w", scanner.Err())
	}
	if cardNumber == "" {
		cardNumber = currentData.CardNumber
	}

	fmt.Fprintf(c.writer, "Текущий срок действия (MM/YY): %s\n", currentData.ExpiryDate)
	fmt.Fprint(c.writer, "Введите новый срок действия (оставьте пустым, чтобы оставить без изменений): ")
	var expiryDate string
	if scanner.Scan() {
		expiryDate = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода срока действия: %w", scanner.Err())
	}
	if expiryDate == "" {
		expiryDate = currentData.ExpiryDate
	}

	fmt.Fprintf(c.writer, "Текущий CVV: %s\n", currentData.CVV)
	fmt.Fprint(c.writer, "Введите новый CVV (оставьте пустым, чтобы оставить без изменений): ")
	var cvv string
	if scanner.Scan() {
		cvv = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода CVV: %w", scanner.Err())
	}
	if cvv == "" {
		cvv = currentData.CVV
	}

	fmt.Fprintf(c.writer, "Текущее имя держателя карты: %s\n", currentData.HolderName)
	fmt.Fprint(c.writer, "Введите новое имя держателя карты (оставьте пустым, чтобы оставить без изменений): ")
	var holderName string
	if scanner.Scan() {
		holderName = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода имени держателя карты: %w", scanner.Err())
	}
	if holderName == "" {
		holderName = currentData.HolderName
	}

	fmt.Fprintf(c.writer, "Текущая метаинформация: %s\n", dataItem.Meta)
	fmt.Fprint(c.writer, "Введите новую метаинформацию (оставьте пустым, чтобы оставить без изменений): ")
	var meta string
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода метаинформации: %w", scanner.Err())
	}
	if meta == "" {
		meta = dataItem.Meta
	}

	updatedData := &entity.BankCardData{
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		CVV:        cvv,
		HolderName: holderName,
	}

	infoBytes, err := json.Marshal(updatedData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации обновленных данных: %w", err)
	}

	updatedDataItem := &datapb.DataItem{
		Id:       dataItem.Id,
		InfoType: dataItem.InfoType,
		Info:     infoBytes,
		Meta:     meta,
	}

	return updatedDataItem, nil
}
