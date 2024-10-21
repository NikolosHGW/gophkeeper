package db

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDBConnector struct {
	mock.Mock
}

func (m *MockDBConnector) Connect(dataSourceName string) (*sqlx.DB, error) {
	args := m.Called(dataSourceName)
	return args.Get(0).(*sqlx.DB), args.Error(1)
}

type MockMigrator struct {
	mock.Mock
}

func (m *MockMigrator) RunMigrations(db *sqlx.DB) error {
	args := m.Called(db)
	return args.Error(0)
}

func TestInitDB_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mockDBConnector := new(MockDBConnector)
	mockMigrator := new(MockMigrator)

	mockDBConnector.On("Connect", "test_db_uri").Return(sqlxDB, nil)
	mockMigrator.On("RunMigrations", sqlxDB).Return(nil)

	result, err := InitDB("test_db_uri", mockDBConnector, mockMigrator)

	assert.NoError(t, err)
	assert.Equal(t, sqlxDB, result)
	mockDBConnector.AssertExpectations(t)
	mockMigrator.AssertExpectations(t)
}

func TestInitDB_ConnectionRetry(t *testing.T) {
	mockDBConnector := new(MockDBConnector)
	mockMigrator := new(MockMigrator)

	retriableError := &pgconn.PgError{
		Code: pgerrcode.ConnectionFailure,
	}

	mockDBConnector.On("Connect", "test_db_uri").Return((*sqlx.DB)(nil), retriableError).Times(3)
	mockDBConnector.On("Connect", "test_db_uri").Return((*sqlx.DB)(nil), errors.New("final connection failure")).Once()

	_, err := InitDB("test_db_uri", mockDBConnector, mockMigrator)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connect to postgres after retries: final connection failure")

	mockDBConnector.AssertNumberOfCalls(t, "Connect", 4)
}

func TestInitDB_MigrationError(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mockDBConnector := new(MockDBConnector)
	mockMigrator := new(MockMigrator)

	mockDBConnector.On("Connect", "test_db_uri").Return(sqlxDB, nil)
	mockMigrator.On("RunMigrations", sqlxDB).Return(errors.New("migration failed"))

	_, err = InitDB("test_db_uri", mockDBConnector, mockMigrator)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration failed")
	mockDBConnector.AssertExpectations(t)
	mockMigrator.AssertExpectations(t)
}

func TestIsRetriableError(t *testing.T) {
	retriableErrors := []string{
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected,
		pgerrcode.LockNotAvailable,
		pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure,
	}

	for _, code := range retriableErrors {
		t.Run("RetriableError_"+code, func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: code}
			assert.True(t, isRetriableError(pgErr), "Expected error to be retriable")
		})
	}

	nonRetriableError := &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	assert.False(t, isRetriableError(nonRetriableError), "Expected error to be non-retriable")

	otherError := errors.New("some other error")
	assert.False(t, isRetriableError(otherError), "Expected non-PgError to be non-retriable")
}
