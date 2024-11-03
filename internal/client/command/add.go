package command

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
)

type dataService interface {
	AddData(ctx context.Context, token string, data *datapb.DataItem) (int32, error)
}

type AddCommand struct {
	dataService dataService
	tokenHolder *entity.TokenHolder
	reader      io.Reader
	writer      io.Writer
}

func NewAddCommand(
	dataService dataService,
	tokenHolder *entity.TokenHolder,
	reader io.Reader,
	writer io.Writer,
) *AddCommand {
	return &AddCommand{
		dataService: dataService,
		tokenHolder: tokenHolder,
		reader:      reader,
		writer:      writer,
	}
}

func (c *AddCommand) Name() string {
	return "add"
}

func (c *AddCommand) Execute() error {
	if c.tokenHolder.Token == "" {
		return fmt.Errorf("вы должны войти в систему")
	}

	scanner := bufio.NewScanner(c.reader)

	fmt.Fprintln(c.writer, "Выберите тип данных для добавления:")
	fmt.Fprintln(c.writer, "1. Login and Password")
	fmt.Fprintln(c.writer, "2. Text")
	fmt.Fprintln(c.writer, "3. Binary File")
	fmt.Fprintln(c.writer, "4. Bank Card")
	fmt.Fprint(c.writer, "Введите номер опции: ")

	var option string
	if scanner.Scan() {
		option = scanner.Text()
	} else {
		return fmt.Errorf("ошибка ввода опции: %w", scanner.Err())
	}

	var dataItem *datapb.DataItem
	var err error

	switch option {
	case "1":
		dataItem, err = c.inputLoginPasswordData(scanner)
	case "2":
		dataItem, err = c.inputTextData(scanner)
	case "3":
		dataItem, err = c.inputBinaryData(scanner)
	case "4":
		dataItem, err = c.inputBankCardData(scanner)
	default:
		fmt.Fprintln(c.writer, "Некорректная опция")
		return nil
	}

	if err != nil {
		return err
	}

	id, err := c.dataService.AddData(context.Background(), c.tokenHolder.Token, dataItem)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных: %w", err)
	}

	fmt.Fprintf(c.writer, "Данные успешно добавлены с ID: %d\n", id)
	return nil
}

func (c *AddCommand) inputLoginPasswordData(scanner *bufio.Scanner) (*datapb.DataItem, error) {
	fmt.Fprint(c.writer, "Введите логин: ")
	var login string
	if scanner.Scan() {
		login = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода логина: %w", scanner.Err())
	}

	fmt.Fprint(c.writer, "Введите пароль: ")
	var password string
	if scanner.Scan() {
		password = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода пароля: %w", scanner.Err())
	}

	fmt.Fprint(c.writer, "Введите URL: ")
	var url string
	if scanner.Scan() {
		url = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода URL: %w", scanner.Err())
	}

	fmt.Fprint(c.writer, "Введите метаинформацию: ")
	var meta string
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода метаинформации: %w", scanner.Err())
	}

	loginPasswordData := &entity.LoginPasswordData{
		Login:    login,
		Password: password,
		URL:      url,
	}

	infoBytes, err := json.Marshal(loginPasswordData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	dataItem := &datapb.DataItem{
		InfoType: "login_password",
		Info:     infoBytes,
		Meta:     meta,
	}

	return dataItem, nil
}

func (c *AddCommand) inputTextData(scanner *bufio.Scanner) (*datapb.DataItem, error) {
	fmt.Fprint(c.writer, "Введите текст: ")
	var text string
	if scanner.Scan() {
		text = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода текста: %w", scanner.Err())
	}

	fmt.Fprint(c.writer, "Введите метаинформацию: ")
	var meta string
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода метаинформации: %w", scanner.Err())
	}

	textData := &entity.TextData{
		Text: text,
	}

	infoBytes, err := json.Marshal(textData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	dataItem := &datapb.DataItem{
		InfoType: "text",
		Info:     infoBytes,
		Meta:     meta,
	}

	return dataItem, nil
}

func (c *AddCommand) inputBinaryData(scanner *bufio.Scanner) (*datapb.DataItem, error) {
	fmt.Fprint(c.writer, "Введите путь к файлу: ")
	var filePath string
	if scanner.Scan() {
		filePath = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода пути к файлу: %w", scanner.Err())
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	fmt.Fprint(c.writer, "Введите метаинформацию: ")
	var meta string
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода метаинформации: %w", scanner.Err())
	}

	fileName := getFileName(filePath)

	binaryData := &entity.BinaryData{
		FileName:    fileName,
		FileContent: fileContent,
	}

	infoBytes, err := json.Marshal(binaryData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	dataItem := &datapb.DataItem{
		InfoType: "binary",
		Info:     infoBytes,
		Meta:     meta,
	}

	return dataItem, nil
}

func (c *AddCommand) inputBankCardData(scanner *bufio.Scanner) (*datapb.DataItem, error) {
	fmt.Fprint(c.writer, "Введите номер карты: ")
	var cardNumber string
	if scanner.Scan() {
		cardNumber = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода номера карты: %w", scanner.Err())
	}

	fmt.Fprint(c.writer, "Введите срок действия (MM/YY): ")
	var expiryDate string
	if scanner.Scan() {
		expiryDate = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода срока действия: %w", scanner.Err())
	}

	fmt.Fprint(c.writer, "Введите CVV: ")
	var cvv string
	if scanner.Scan() {
		cvv = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода CVV: %w", scanner.Err())
	}

	fmt.Fprint(c.writer, "Введите имя держателя карты: ")
	var holderName string
	if scanner.Scan() {
		holderName = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода имени держателя карты: %w", scanner.Err())
	}

	fmt.Fprint(c.writer, "Введите метаинформацию: ")
	var meta string
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return nil, fmt.Errorf("ошибка ввода метаинформации: %w", scanner.Err())
	}

	bankCardData := &entity.BankCardData{
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		CVV:        cvv,
		HolderName: holderName,
	}

	infoBytes, err := json.Marshal(bankCardData)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	dataItem := &datapb.DataItem{
		InfoType: "bank_card",
		Info:     infoBytes,
		Meta:     meta,
	}

	return dataItem, nil
}

func getFileName(filePath string) string {
	_, fileName := filepath.Split(filePath)
	return fileName
}
