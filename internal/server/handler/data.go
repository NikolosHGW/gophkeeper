package handler

import (
	"context"
	"errors"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/contextkey"
	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type dataService interface {
	AddData(ctx context.Context, userID int, data *entity.UserData) (int, error)
	GetDataByID(ctx context.Context, userID, dataID int) (*entity.UserData, error)
	UpdateData(ctx context.Context, userID int, data *entity.UserData) error
	DeleteData(ctx context.Context, userID, dataID int) error
}

type DataServer struct {
	datapb.UnimplementedDataServiceServer
	dataService dataService
	logger      logger.CustomLogger
}

func NewDataServer(dataService dataService, logger logger.CustomLogger) *DataServer {
	return &DataServer{
		dataService: dataService,
		logger:      logger,
	}
}

func (h *DataServer) AddData(ctx context.Context, req *datapb.AddDataRequest) (*datapb.AddDataResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось получить userID из контекста")
	}

	data := &entity.UserData{
		InfoType: req.Data.InfoType,
		Info:     req.Data.Info,
		Meta:     req.Data.Meta,
	}

	id, err := h.dataService.AddData(ctx, userID, data)
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка при добавлении данных")
	}

	return &datapb.AddDataResponse{Id: int32(id)}, nil
}

func (h *DataServer) GetData(ctx context.Context, req *datapb.GetDataRequest) (*datapb.GetDataResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	data, err := h.dataService.GetDataByID(ctx, userID, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "данные не найдены")
	}

	return &datapb.GetDataResponse{
		Data: &datapb.DataItem{
			Id:       int32(data.ID),
			InfoType: data.InfoType,
			Info:     data.Info,
			Meta:     data.Meta,
			Created:  timestamppb.New(data.Created),
		},
	}, nil
}

func (h *DataServer) UpdateData(ctx context.Context, req *datapb.UpdateDataRequest) (*datapb.UpdateDataResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.logger.LogInfo("Не удалось получить userID из контекста", err)
		return nil, status.Error(codes.Internal, "не удалось получить userID из контекста")
	}

	data := &entity.UserData{
		ID:       int(req.Data.Id),
		UserID:   userID,
		InfoType: req.Data.InfoType,
		Info:     req.Data.Info,
		Meta:     req.Data.Meta,
		Created:  req.Data.Created.AsTime(),
	}

	err = h.dataService.UpdateData(ctx, userID, data)
	if err != nil {
		h.logger.LogInfo("Ошибка при обновлении данных", err)
		return nil, status.Error(codes.Internal, "ошибка при обновлении данных")
	}

	return &datapb.UpdateDataResponse{}, nil
}

func (h *DataServer) DeleteData(ctx context.Context, req *datapb.DeleteDataRequest) (*datapb.DeleteDataResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.logger.LogInfo("Не удалось получить userID из контекста", err)
		return nil, status.Error(codes.Internal, "не удалось получить userID из контекста")
	}

	err = h.dataService.DeleteData(ctx, userID, int(req.Id))
	if err != nil {
		h.logger.LogInfo("Ошибка при удалении данных", err)
		return nil, status.Error(codes.Internal, "ошибка при удалении данных")
	}

	return &datapb.DeleteDataResponse{}, nil
}

func getUserIDFromContext(ctx context.Context) (int, error) {
	userIDValue := ctx.Value(contextkey.UserIDKey)
	if userIDValue == nil {
		return 0, errors.New("userID не найден в контексте")
	}
	userID, ok := userIDValue.(int)
	if !ok {
		return 0, errors.New("userID имеет неверный тип")
	}
	return userID, nil
}
