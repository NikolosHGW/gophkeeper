package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/internal/server/helper"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type mockLogger struct{}

func (l *mockLogger) LogInfo(message string, err error) {}

func TestUser_Save_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockLogger := new(mockLogger)
	repo := NewUser(sqlxDB, mockLogger)

	user := &entity.User{
		Login:    "testuser",
		Password: "password123",
	}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("testuser", "password123").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Save(context.Background(), user)

	assert.NoError(t, err)
	assert.Equal(t, int(1), user.ID)
}

func TestUser_Save_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockLogger := new(mockLogger)
	repo := NewUser(sqlxDB, mockLogger)

	user := &entity.User{
		Login:    "testuser",
		Password: "password123",
	}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("testuser", "password123").
		WillReturnError(errors.New("some error"))

	err = repo.Save(context.Background(), user)

	assert.Error(t, err)
	assert.Equal(t, helper.ErrInternalServer, err)
}

func TestUser_ExistsByLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockLogger := new(mockLogger)
	repo := NewUser(sqlxDB, mockLogger)

	login := "testuser"

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(login).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ExistsByLogin(context.Background(), login)

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUser_ExistsByLogin_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockLogger := new(mockLogger)
	repo := NewUser(sqlxDB, mockLogger)

	login := "testuser"

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(login).
		WillReturnError(errors.New("some error"))

	exists, err := repo.ExistsByLogin(context.Background(), login)

	assert.Error(t, err)
	assert.False(t, exists)
	assert.Equal(t, helper.ErrInternalServer, err)
}

func TestUser_User_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockLogger := new(mockLogger)
	repo := NewUser(sqlxDB, mockLogger)

	login := "testuser"
	expectedUser := &entity.User{
		ID:       1,
		Login:    "testuser",
		Password: "password123",
	}

	rows := sqlmock.NewRows([]string{"id", "login", "password"}).
		AddRow(expectedUser.ID, expectedUser.Login, expectedUser.Password)

	mock.ExpectQuery("SELECT id, login, password FROM users WHERE login = \\$1").
		WithArgs(login).
		WillReturnRows(rows)

	user, err := repo.User(context.Background(), login)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestUser_User_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockLogger := new(mockLogger)
	repo := NewUser(sqlxDB, mockLogger)

	login := "nonexistentuser"

	mock.ExpectQuery("SELECT id, login, password FROM users WHERE login = \\$1").
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.User(context.Background(), login)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, helper.ErrInvalidCredentials, err)
}

func TestUser_User_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	mockLogger := new(mockLogger)
	repo := NewUser(sqlxDB, mockLogger)

	login := "testuser"

	mock.ExpectQuery("SELECT id, login, password FROM users WHERE login = \\$1").
		WithArgs(login).
		WillReturnError(errors.New("database error"))

	user, err := repo.User(context.Background(), login)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.EqualError(t, err, "ошибка при поиске пользователя")
}
