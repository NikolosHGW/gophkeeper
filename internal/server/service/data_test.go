package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type DataRepoMock struct {
	mock.Mock
}

func (m *DataRepoMock) AddData(ctx context.Context, data *entity.UserData) (int, error) {
	args := m.Called(ctx, data)
	return args.Int(0), args.Error(1)
}

func (m *DataRepoMock) GetDataByID(ctx context.Context, userID, dataID int) (*entity.UserData, error) {
	args := m.Called(ctx, userID, dataID)
	return args.Get(0).(*entity.UserData), args.Error(1)
}

func (m *DataRepoMock) UpdateData(ctx context.Context, data *entity.UserData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *DataRepoMock) DeleteData(ctx context.Context, userID, dataID int) error {
	args := m.Called(ctx, userID, dataID)
	return args.Error(0)
}

func (m *DataRepoMock) ListData(ctx context.Context, userID int, infoType string) ([]*entity.UserData, error) {
	args := m.Called(ctx, userID, infoType)
	return args.Get(0).([]*entity.UserData), args.Error(1)
}

func TestDataService_AddData(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptionService := NewEncryptionService(key)

	dataRepoMock := new(DataRepoMock)
	dataService := NewDataService(dataRepoMock, encryptionService)

	ctx := context.Background()
	userID := 1
	data := &entity.UserData{
		InfoType: "text",
		Info:     "секретная информация",
		Meta:     "метаданные",
	}

	dataRepoMock.On("AddData", ctx, mock.AnythingOfType("*entity.UserData")).Return(1, nil).Run(func(args mock.Arguments) {
		argData := args.Get(1).(*entity.UserData)
		assert.NotEqual(t, "секретная информация", argData.Info)
		assert.NotEqual(t, "метаданные", argData.Meta)
	})

	id, err := dataService.AddData(ctx, userID, data)
	assert.NoError(t, err)
	assert.Equal(t, 1, id)

	dataRepoMock.AssertExpectations(t)
}

func TestDataService_GetDataByID(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptionService := NewEncryptionService(key)

	dataRepoMock := new(DataRepoMock)
	dataService := NewDataService(dataRepoMock, encryptionService)

	ctx := context.Background()
	userID := 1
	dataID := 1

	encryptedInfo, _ := encryptionService.Encrypt("секретная информация")
	encryptedMeta, _ := encryptionService.Encrypt("метаданные")

	storedData := &entity.UserData{
		ID:       dataID,
		UserID:   userID,
		InfoType: "text",
		Info:     encryptedInfo,
		Meta:     encryptedMeta,
	}

	dataRepoMock.On("GetDataByID", ctx, userID, dataID).Return(storedData, nil)

	data, err := dataService.GetDataByID(ctx, userID, dataID)
	assert.NoError(t, err)
	assert.Equal(t, "секретная информация", data.Info)
	assert.Equal(t, "метаданные", data.Meta)

	dataRepoMock.AssertExpectations(t)
}

func TestDataService_UpdateData(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptionService := NewEncryptionService(key)

	dataRepoMock := new(DataRepoMock)
	dataService := NewDataService(dataRepoMock, encryptionService)

	ctx := context.Background()
	userID := 1
	data := &entity.UserData{
		ID:       1,
		InfoType: "text",
		Info:     "обновленная информация",
		Meta:     "обновленные метаданные",
	}

	dataRepoMock.On("UpdateData", ctx, mock.AnythingOfType("*entity.UserData")).Return(nil).Run(func(args mock.Arguments) {
		argData := args.Get(1).(*entity.UserData)
		assert.NotEqual(t, "обновленная информация", argData.Info)
		assert.NotEqual(t, "обновленные метаданные", argData.Meta)
	})

	err := dataService.UpdateData(ctx, userID, data)
	assert.NoError(t, err)

	dataRepoMock.AssertExpectations(t)
}

func TestDataService_DeleteData(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptionService := NewEncryptionService(key)

	dataRepoMock := new(DataRepoMock)
	dataService := NewDataService(dataRepoMock, encryptionService)

	ctx := context.Background()
	userID := 1
	dataID := 1

	dataRepoMock.On("DeleteData", ctx, userID, dataID).Return(nil)

	err := dataService.DeleteData(ctx, userID, dataID)
	assert.NoError(t, err)

	dataRepoMock.AssertExpectations(t)
}

func TestDataService_GetDataByID_DecryptionError(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptionService := NewEncryptionService(key)

	dataRepoMock := new(DataRepoMock)
	dataService := NewDataService(dataRepoMock, encryptionService)

	ctx := context.Background()
	userID := 1
	dataID := 1

	storedData := &entity.UserData{
		ID:       dataID,
		UserID:   userID,
		InfoType: "text",
		Info:     "некорректные данные",
		Meta:     "некорректные метаданные",
	}

	dataRepoMock.On("GetDataByID", ctx, userID, dataID).Return(storedData, nil)

	_, err := dataService.GetDataByID(ctx, userID, dataID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка расшифровки")

	dataRepoMock.AssertExpectations(t)
}

func TestDataService_ListData(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptionService := NewEncryptionService(key)

	dataRepoMock := new(DataRepoMock)
	dataService := NewDataService(dataRepoMock, encryptionService)

	ctx := context.Background()
	userID := 1
	infoType := "text"

	encryptedMeta1, _ := encryptionService.Encrypt("метаданные1")
	encryptedMeta2, _ := encryptionService.Encrypt("метаданные2")

	storedData := []*entity.UserData{
		{
			ID:       1,
			UserID:   userID,
			InfoType: "text",
			Meta:     string(encryptedMeta1),
			Created:  time.Now(),
		},
		{
			ID:       2,
			UserID:   userID,
			InfoType: "text",
			Meta:     string(encryptedMeta2),
			Created:  time.Now(),
		},
	}

	dataRepoMock.On("ListData", ctx, userID, infoType).Return(storedData, nil)

	dataItems, err := dataService.ListData(ctx, userID, infoType)
	assert.NoError(t, err)
	assert.Len(t, dataItems, 2)

	assert.Equal(t, "метаданные1", dataItems[0].Meta)
	assert.Equal(t, "метаданные2", dataItems[1].Meta)

	dataRepoMock.AssertExpectations(t)
}

func TestDataService_ListData_Error(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptionService := NewEncryptionService(key)

	dataRepoMock := new(DataRepoMock)
	dataService := NewDataService(dataRepoMock, encryptionService)

	ctx := context.Background()
	userID := 1
	infoType := "text"

	dataRepoMock.On("ListData", ctx, userID, infoType).Return(([]*entity.UserData)(nil), fmt.Errorf("database error"))

	_, err := dataService.ListData(ctx, userID, infoType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка получения данных из репозитория")

	dataRepoMock.AssertExpectations(t)
}
