package service

import (
	"testing"

	pb "github.com/NikolosHGW/goph-keeper/api/registerpb"
	"golang.org/x/crypto/bcrypt"
)

type mockLogger struct{}

func (l *mockLogger) LogInfo(message string, err error) {}

func TestRegister_CreateUser_Success(t *testing.T) {
	mockLogger := &mockLogger{}
	reg := NewRegister(mockLogger)

	req := &pb.RegisterUserRequest{
		Login:    "testuser",
		Password: "password123",
	}

	user, err := reg.CreateUser(req)

	if err != nil {
		t.Fatalf("Ожидалось отсутствие ошибки, но получена: %v", err)
	}

	if user == nil {
		t.Fatal("Ожидался пользователь, но получен nil")
	}

	if user.Login != req.Login {
		t.Errorf("Ожидался логин %s, но получен %s", req.Login, user.Login)
	}

	if user.Password == req.Password {
		t.Errorf("Ожидалось, что пароль будет хеширован и не совпадет с исходным")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		t.Errorf("Хеш пароля не соответствует исходному паролю: %v", err)
	}
}

func TestRegister_CreateUser_HashError(t *testing.T) {
	mockLogger := &mockLogger{}
	reg := NewRegister(mockLogger)

	longPassword := make([]byte, 1<<24)
	for i := range longPassword {
		longPassword[i] = 'a'
	}

	req := &pb.RegisterUserRequest{
		Login:    "testuser",
		Password: string(longPassword),
	}

	user, err := reg.CreateUser(req)

	if err == nil {
		t.Fatal("Ожидалась ошибка хеширования пароля, но ошибки нет")
	}

	if user != nil {
		t.Fatal("Ожидался nil пользователь при ошибке хеширования")
	}
}
