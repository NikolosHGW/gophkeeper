package command

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
)

type listDataService interface {
	ListData(ctx context.Context, token string, filter *entity.DataFilter) ([]*datapb.DataItem, error)
}

type ListCommand struct {
	dataService listDataService
	tokenHolder *entity.TokenHolder
	reader      io.Reader
	writer      io.Writer
}

func NewListCommand(
	dataService listDataService,
	tokenHolder *entity.TokenHolder,
	reader io.Reader,
	writer io.Writer,
) *ListCommand {
	return &ListCommand{
		dataService: dataService,
		tokenHolder: tokenHolder,
		reader:      reader,
		writer:      writer,
	}
}

func (c *ListCommand) Name() string {
	return "list"
}

func (c *ListCommand) Execute() error {
	if c.tokenHolder.Token == "" {
		return fmt.Errorf("вы должны войти в систему")
	}

	scanner := bufio.NewScanner(c.reader)

	fmt.Fprint(c.writer, "Введите тип данных для фильтрации (оставьте пустым для всех типов): ")
	var infoType string
	if scanner.Scan() {
		infoType = scanner.Text()
	} else {
		return fmt.Errorf("ошибка ввода типа данных: %w", scanner.Err())
	}

	filter := &entity.DataFilter{InfoType: infoType}
	dataItems, err := c.dataService.ListData(context.Background(), c.tokenHolder.Token, filter)
	if err != nil {
		return fmt.Errorf("ошибка получения списка данных: %w", err)
	}

	if len(dataItems) == 0 {
		fmt.Fprintln(c.writer, "Данные не найдены.")
		return nil
	}

	fmt.Fprintln(c.writer, "Список данных:")
	for _, item := range dataItems {
		fmt.Fprintf(c.writer, "ID: %d, Тип: %s, Мета: %s, Дата создания: %s\n",
			item.Id, item.InfoType, item.Meta, item.Created.AsTime().Format("2006-01-02 15:04:05"))
	}

	return nil
}
