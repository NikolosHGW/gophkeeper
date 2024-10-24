package command

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
)

type deleteDataService interface {
	DeleteData(ctx context.Context, token string, id int32) error
}

type DeleteCommand struct {
	dataService deleteDataService
	tokenHolder *entity.TokenHolder
	reader      io.Reader
	writer      io.Writer
}

func NewDeleteCommand(
	dataService deleteDataService,
	tokenHolder *entity.TokenHolder,
	reader io.Reader,
	writer io.Writer,
) *DeleteCommand {
	return &DeleteCommand{
		dataService: dataService,
		tokenHolder: tokenHolder,
		reader:      reader,
		writer:      writer,
	}
}

func (c *DeleteCommand) Name() string {
	return "delete"
}

func (c *DeleteCommand) Execute() error {
	if c.tokenHolder.Token == "" {
		return fmt.Errorf("вы должны войти в систему")
	}

	_, err := fmt.Fprint(c.writer, "Введите ID данных: ")
	if err != nil {
		return fmt.Errorf("ошибка вывода запроса на ввод ID: %w", err)
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

	err = c.dataService.DeleteData(context.Background(), c.tokenHolder.Token, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления данных: %w", err)
	}

	_, err = fmt.Fprintln(c.writer, "Данные успешно удалены.")
	if err != nil {
		return fmt.Errorf("ошибка вывода результата: %w", err)
	}

	return nil
}
