package command

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
)

type service interface {
	Login(ctx context.Context, login, password string) (string, error)
}

type LoginCommand struct {
	authService service
	tokenHolder *entity.TokenHolder
	reader      io.Reader
	writer      io.Writer
}

func NewLoginCommand(
	authService service,
	tokenHolder *entity.TokenHolder,
	reader io.Reader,
	writer io.Writer,
) *LoginCommand {
	return &LoginCommand{
		authService: authService,
		tokenHolder: tokenHolder,
		reader:      reader,
		writer:      writer,
	}
}

func (c *LoginCommand) Name() string {
	return "login"
}

func (c *LoginCommand) Execute() error {
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

	token, err := c.authService.Login(context.Background(), login, password)
	if err != nil {
		return fmt.Errorf("ошибка входа: %w", err)
	}

	c.tokenHolder.Token = token
	fmt.Println("Вход выполнен успешно.")
	return nil
}
