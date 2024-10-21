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

	_, err = fmt.Fprintf(c.writer, "Текущий тип (%s): ", dataItem.InfoType)
	if err != nil {
		return fmt.Errorf("ошибка вывода текущего типа: %w", err)
	}
	if !scanner.Scan() {
		return fmt.Errorf("ошибка ввода текущего типа информации")
	}
	infoType := scanner.Text()
	if infoType == "" {
		infoType = dataItem.InfoType
	}

	_, err = fmt.Fprintf(c.writer, "Текущие данные (%s): ", dataItem.Info)
	if err != nil {
		return fmt.Errorf("ошибка вывода текущих данных: %w", err)
	}
	if !scanner.Scan() {
		return fmt.Errorf("ошибка ввода текущих данных")
	}
	info := scanner.Text()
	if info == "" {
		info = dataItem.Info
	}

	_, err = fmt.Fprintf(c.writer, "Текущая мета (%s): ", dataItem.Meta)
	if err != nil {
		return fmt.Errorf("ошибка вывода текущей метаинформации: %w", err)
	}
	if !scanner.Scan() {
		return fmt.Errorf("ошибка ввода текущей метаинформации")
	}
	meta := scanner.Text()
	if meta == "" {
		meta = dataItem.Meta
	}

	updatedData := &datapb.DataItem{
		Id:       id,
		InfoType: infoType,
		Info:     info,
		Meta:     meta,
		Created:  dataItem.Created,
	}

	err = c.dataService.UpdateData(context.Background(), c.tokenHolder.Token, updatedData)
	if err != nil {
		return fmt.Errorf("ошибка обновления данных: %w", err)
	}

	_, err = fmt.Fprintln(c.writer, "Данные успешно обновлены.")
	if err != nil {
		return fmt.Errorf("ошибка вывода результата: %w", err)
	}

	return nil
}
