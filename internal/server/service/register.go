package service

import (
	pb "github.com/NikolosHGW/goph-keeper/api/registerpb"
	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type register struct {
	log logger.CustomLogger
}

func NewRegister(log logger.CustomLogger) *register {
	return &register{
		log: log,
	}
}

func (u *register) CreateUser(req *pb.RegisterUserRequest) (*entity.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		u.log.LogInfo("ошибка при хэшировании пароля: ", err)
		return nil, helper.ErrInternalServer
	}

	user := &entity.User{
		Login:    req.Login,
		Password: string(passwordHash),
	}

	return user, nil
}
