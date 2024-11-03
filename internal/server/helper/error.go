package helper

import "errors"

var (
	ErrLoginAlreadyExists = errors.New("логин уже существует")
	ErrInvalidCredentials = errors.New("неверная пара логин/пароль")
	ErrInternalServer     = errors.New("внутренняя ошибка сервера")
)
