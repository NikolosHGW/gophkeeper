package command

import (
	"bufio"
	"context"
	"fmt"
	"io"
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

	return nil
}
