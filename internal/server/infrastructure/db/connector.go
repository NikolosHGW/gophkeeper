package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type DBConnector struct{}

func (d *DBConnector) Connect(dataSourceName string) (*sqlx.DB, error) {
	driver, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("ошибка sqlx.Connect: %w", err)
	}

	return driver, nil
}
