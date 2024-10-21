package command

import (
	"bufio"
	"context"
	"fmt"
	"io"

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

	var infoType, info, meta string
	_, err := fmt.Fprint(c.writer, "Введите тип информации (login_password, text, binary, bank_card): ")
	if err != nil {
		return fmt.Errorf("ошибка stdin тип информации: %w", err)
	}
	scanner := bufio.NewScanner(c.reader)
	if scanner.Scan() {
		infoType = scanner.Text()
	} else {
		return fmt.Errorf("ошибка ввода типа информации: unexpected EOF")
	}

	_, err = fmt.Fprint(c.writer, "Введите данные: ")
	if err != nil {
		return fmt.Errorf("ошибка stdin ввода данных: %w", err)
	}
	if scanner.Scan() {
		info = scanner.Text()
	} else {
		return fmt.Errorf("ошибка ввода пароля: unexpected EOF")
	}

	_, err = fmt.Fprint(c.writer, "Введите метаинформацию: ")
	if err != nil {
		return fmt.Errorf("ошибка stdin ввода метаинформации: %w", err)
	}
	if scanner.Scan() {
		meta = scanner.Text()
	} else {
		return fmt.Errorf("ошибка ввода метаинформации: unexpected EOF")
	}

	dataItem := &datapb.DataItem{
		InfoType: infoType,
		Info:     info,
		Meta:     meta,
	}

	id, err := c.dataService.AddData(context.Background(), c.tokenHolder.Token, dataItem)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных: %w", err)
	}

	_, err = fmt.Fprintf(c.writer, "Данные успешно добавлены с ID: %d\n", id)
	if err != nil {
		return fmt.Errorf("ошибка вывода результата: %w", err)
	}

	return nil
}
