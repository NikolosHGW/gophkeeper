package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type storager interface {
	QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type User struct {
	db     storager
	logger logger.CustomLogger
}

func NewUser(db storager, logger logger.CustomLogger) *User {
	return &User{db: db, logger: logger}
}

func (r *User) Save(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowxContext(ctx, query, user.Login, user.Password).Scan(&user.ID)
	if err != nil {
		r.logger.LogInfo("ошибка при сохранении пользователя", err)
		return helper.ErrInternalServer
	}

	return nil
}

func (r *User) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE login=$1)`
	err := r.db.QueryRowxContext(ctx, query, login).Scan(&exists)
	if err != nil {
		r.logger.LogInfo("не получилось записать результат запроса в переменную", err)
		return false, helper.ErrInternalServer
	}
	return exists, nil
}

func (r *User) User(ctx context.Context, login string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, login, password FROM users WHERE login = $1`
	err := r.db.GetContext(ctx, &user, query, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, helper.ErrInvalidCredentials
		}

		r.logger.LogInfo("ошибка при поиске пользователя: ", err)

		return nil, fmt.Errorf("ошибка при поиске пользователя")
	}
	return &user, nil
}
