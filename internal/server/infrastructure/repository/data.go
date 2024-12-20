package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
)

type dataStorager interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type dataRepository struct {
	db     dataStorager
	logger logger.CustomLogger
}

// NewDataRepository - конструктор data repo.
func NewDataRepository(db dataStorager, logger logger.CustomLogger) *dataRepository {
	return &dataRepository{db: db, logger: logger}
}

func (r *dataRepository) AddData(ctx context.Context, data *entity.UserData) (int, error) {
	query := `
        INSERT INTO user_data (user_id, info_type, info, meta, created)
        VALUES ($1, $2, $3, $4, NOW())
        RETURNING id
    `
	var id int
	err := r.db.QueryRowContext(ctx, query, data.UserID, data.InfoType, data.Info, data.Meta).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *dataRepository) GetDataByID(ctx context.Context, userID, dataID int) (*entity.UserData, error) {
	query := `
        SELECT id, user_id, info_type, info, meta, created
        FROM user_data
        WHERE id = $1 AND user_id = $2
    `
	row := r.db.QueryRowContext(ctx, query, dataID, userID)
	data := &entity.UserData{}
	err := row.Scan(&data.ID, &data.UserID, &data.InfoType, &data.Info, &data.Meta, &data.Created)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *dataRepository) UpdateData(ctx context.Context, data *entity.UserData) error {
	query := `
        UPDATE user_data
        SET info_type = $1, info = $2, meta = $3
        WHERE id = $4 AND user_id = $5
    `
	_, err := r.db.ExecContext(ctx, query, data.InfoType, data.Info, data.Meta, data.ID, data.UserID)
	return err
}

func (r *dataRepository) DeleteData(ctx context.Context, userID, dataID int) error {
	query := `
        DELETE FROM user_data
        WHERE id = $1 AND user_id = $2
    `
	_, err := r.db.ExecContext(ctx, query, dataID, userID)
	return err
}

func (r *dataRepository) ListData(ctx context.Context, userID int, infoType string) ([]*entity.UserData, error) {
	var dataItems []*entity.UserData

	query := `SELECT id, user_id, info_type, info, meta, created FROM user_data WHERE user_id = $1`
	args := []interface{}{userID}

	if infoType != "" {
		query += ` AND info_type = $2`
		args = append(args, infoType)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса к базе данных: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var data entity.UserData
		err := rows.Scan(&data.ID, &data.UserID, &data.InfoType, &data.Info, &data.Meta, &data.Created)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения данных из базы данных: %w", err)
		}
		dataItems = append(dataItems, &data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", err)
	}

	return dataItems, nil
}
