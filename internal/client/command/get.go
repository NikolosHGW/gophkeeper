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

type getDataService interface {
	GetData(ctx context.Context, token string, id int32) (*datapb.DataItem, error)
}

type GetCommand struct {
	dataService getDataService
	tokenHolder *entity.TokenHolder
	reader      io.Reader
	writer      io.Writer
}

func NewGetCommand(
	dataService getDataService,
	tokenHolder *entity.TokenHolder,
	reader io.Reader,
	writer io.Writer,
) *GetCommand {
	return &GetCommand{
		dataService: dataService,
		tokenHolder: tokenHolder,
		reader:      reader,
		writer:      writer,
	}
}

func (c *GetCommand) Name() string {
	return "get"
}

func (c *GetCommand) Execute() error {
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

	fmt.Fprintf(c.writer, "ID: %d\n", dataItem.Id)
	fmt.Fprintf(c.writer, "Тип: %s\n", dataItem.InfoType)
	fmt.Fprintf(c.writer, "Данные: %s\n", dataItem.Info)
	fmt.Fprintf(c.writer, "Мета: %s\n", dataItem.Meta)
	fmt.Fprintf(c.writer, "Создано: %s\n", dataItem.Created.AsTime())

	switch dataItem.InfoType {
	case "login_password":
		var loginPasswordData entity.LoginPasswordData
		if err := json.Unmarshal(dataItem.Info, &loginPasswordData); err != nil {
			return fmt.Errorf("ошибка десериализации данных: %w", err)
		}
		fmt.Fprintf(c.writer, "Логин: %s\n", loginPasswordData.Login)
		fmt.Fprintf(c.writer, "Пароль: %s\n", loginPasswordData.Password)
		fmt.Fprintf(c.writer, "URL: %s\n", loginPasswordData.URL)
	case "text":
		var textData entity.TextData
		if err := json.Unmarshal(dataItem.Info, &textData); err != nil {
			return fmt.Errorf("ошибка десериализации данных: %w", err)
		}
		fmt.Fprintf(c.writer, "Текст: %s\n", textData.Text)
	case "binary":
		var binaryData entity.BinaryData
		if err := json.Unmarshal(dataItem.Info, &binaryData); err != nil {
			return fmt.Errorf("ошибка десериализации данных: %w", err)
		}
		err = os.WriteFile(binaryData.FileName, binaryData.FileContent, 0644)
		if err != nil {
			return fmt.Errorf("ошибка сохранения файла: %w", err)
		}
		fmt.Fprintf(c.writer, "Бинарные данные сохранены в файл: %s\n", binaryData.FileName)
	case "bank_card":
		var bankCardData entity.BankCardData
		if err := json.Unmarshal(dataItem.Info, &bankCardData); err != nil {
			return fmt.Errorf("ошибка десериализации данных: %w", err)
		}
		fmt.Fprintf(c.writer, "Номер карты: %s\n", bankCardData.CardNumber)
		fmt.Fprintf(c.writer, "Срок действия: %s\n", bankCardData.ExpiryDate)
		fmt.Fprintf(c.writer, "CVV: %s\n", bankCardData.CVV)
		fmt.Fprintf(c.writer, "Имя держателя: %s\n", bankCardData.HolderName)
	default:
		fmt.Fprintln(c.writer, "Неизвестный тип данных")
	}

	return nil
}
