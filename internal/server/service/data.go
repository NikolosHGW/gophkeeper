package service

import (
	"context"
	"fmt"

	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
)

type dataRepo interface {
	AddData(ctx context.Context, data *entity.UserData) (int, error)
	GetDataByID(ctx context.Context, userID, dataID int) (*entity.UserData, error)
	UpdateData(ctx context.Context, data *entity.UserData) error
	DeleteData(ctx context.Context, userID, dataID int) error
}

type dataService struct {
	dataRepo          dataRepo
	encryptionService *EncryptionService
}

// NewDataService - конструктор data service.
func NewDataService(dataRepo dataRepo, encryptionService *EncryptionService) *dataService {
	return &dataService{
		dataRepo:          dataRepo,
		encryptionService: encryptionService,
	}
}

func (s *dataService) AddData(ctx context.Context, userID int, data *entity.UserData) (int, error) {
	data.UserID = userID

	// Шифруем поля data.Info и data.Meta
	encryptedInfo, err := s.encryptionService.Encrypt(data.Info)
	if err != nil {
		return 0, fmt.Errorf("ошибка шифрования Info: %w", err)
	}
	data.Info = encryptedInfo

	encryptedMeta, err := s.encryptionService.Encrypt(data.Meta)
	if err != nil {
		return 0, fmt.Errorf("ошибка шифрования Meta: %w", err)
	}
	data.Meta = encryptedMeta

	return s.dataRepo.AddData(ctx, data)
}

func (s *dataService) GetDataByID(ctx context.Context, userID, dataID int) (*entity.UserData, error) {
	data, err := s.dataRepo.GetDataByID(ctx, userID, dataID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных из репозитория: %w", err)
	}

	// Расшифровываем поля data.Info и data.Meta
	decryptedInfo, err := s.encryptionService.Decrypt(data.Info)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки Info: %w", err)
	}
	data.Info = decryptedInfo

	decryptedMeta, err := s.encryptionService.Decrypt(data.Meta)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки Meta: %w", err)
	}
	data.Meta = decryptedMeta

	return data, nil
}

func (s *dataService) UpdateData(ctx context.Context, userID int, data *entity.UserData) error {
	data.UserID = userID

	// Шифруем поля перед обновлением
	encryptedInfo, err := s.encryptionService.Encrypt(data.Info)
	if err != nil {
		return fmt.Errorf("ошибка шифрования Info: %w", err)
	}
	data.Info = encryptedInfo

	encryptedMeta, err := s.encryptionService.Encrypt(data.Meta)
	if err != nil {
		return fmt.Errorf("ошибка шифрования Meta: %w", err)
	}
	data.Meta = encryptedMeta

	return s.dataRepo.UpdateData(ctx, data)
}

func (s *dataService) DeleteData(ctx context.Context, userID, dataID int) error {
	return s.dataRepo.DeleteData(ctx, userID, dataID)
}
