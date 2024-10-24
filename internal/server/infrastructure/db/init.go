package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type dbConnector interface {
	Connect(dataSourceName string) (*sqlx.DB, error)
}

type migrator interface {
	RunMigrations(db *sqlx.DB) error
}

// InitDB инициализация базы данных, поднятия миграций.
func InitDB(dataSourceName string, dbConnector dbConnector, migrator migrator) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	retryIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for i := 0; i < len(retryIntervals)+1; i++ {
		db, err = dbConnector.Connect(dataSourceName)
		if err == nil {
			break
		}

		if i < len(retryIntervals) {
			if isRetriableError(err) {
				time.Sleep(retryIntervals[i])
				continue
			}
			return nil, fmt.Errorf("connect to postgres: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("connect to postgres after retries: %w", err)
	}

	err = migrator.RunMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("error run migrations: %w", err)
	}

	return db, nil
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected,
			pgerrcode.LockNotAvailable,
			pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure:
			return true
		}
	}
	return false
}
