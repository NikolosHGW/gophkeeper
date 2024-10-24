package command

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
)

type authService interface {
	Register(ctx context.Context, login, password string) (string, error)
}

type RegisterCommand struct {
	authService authService
	tokenHolder *entity.TokenHolder
	reader      io.Reader
	writer      io.Writer
}

func NewRegisterCommand(
	authService authService,
	tokenHolder *entity.TokenHolder,
	reader io.Reader,
	writer io.Writer,
) *RegisterCommand {
	return &RegisterCommand{
		authService: authService,
		tokenHolder: tokenHolder,
		reader:      reader,
		writer:      writer,
	}
}

func (c *RegisterCommand) Name() string {
	return "register"
}

func (c *RegisterCommand) Execute() error {
	var login, password string
	_, err := fmt.Fprint(c.writer, "Введите login: ")
	if err != nil {
		return fmt.Errorf("ошибка stdin login: %w", err)
	}
	scanner := bufio.NewScanner(c.reader)
	if scanner.Scan() {
		login = scanner.Text()
	} else {
		return fmt.Errorf("ошибка ввода логина: %w", scanner.Err())
	}

	_, err = fmt.Fprint(c.writer, "Введите password: ")
	if err != nil {
		return fmt.Errorf("ошибка stdin password: %w", err)
	}
	if scanner.Scan() {
		password = scanner.Text()
	} else {
		return fmt.Errorf("ошибка ввода пароля: %w", scanner.Err())
	}

	token, err := c.authService.Register(context.Background(), login, password)
	if err != nil {
		return fmt.Errorf("ошибка регистрации: %w", err)
	}

	c.tokenHolder.Token = token
	_, err = fmt.Fprintln(c.writer, "Регистрация прошла успешно.")
	if err != nil {
		return fmt.Errorf("ошибка Fprintln : %w", err)
	}
	return nil
}
