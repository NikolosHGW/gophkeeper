package service

import (
	"context"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
	"google.golang.org/grpc/metadata"
)

type dataService struct {
	client datapb.DataServiceClient
	logger logger.CustomLogger
}

func NewDataService(grpcClient *GRPCClient, logger logger.CustomLogger) *dataService {
	return &dataService{client: grpcClient.DataClient, logger: logger}
}

func (s *dataService) AddData(ctx context.Context, token string, data *datapb.DataItem) (int32, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

	req := &datapb.AddDataRequest{Data: data}
	res, err := s.client.AddData(ctx, req)
	if err != nil {
		return 0, err
	}
	return res.Id, nil
}

func (s *dataService) GetData(ctx context.Context, token string, id int32) (*datapb.DataItem, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

	req := &datapb.GetDataRequest{Id: id}
	res, err := s.client.GetData(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (s *dataService) UpdateData(ctx context.Context, token string, data *datapb.DataItem) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

	req := &datapb.UpdateDataRequest{Data: data}
	_, err := s.client.UpdateData(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *dataService) DeleteData(ctx context.Context, token string, id int32) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

	req := &datapb.DeleteDataRequest{Id: id}
	_, err := s.client.DeleteData(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
