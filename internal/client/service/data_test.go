package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MockDataServiceClient struct {
	mock.Mock
}

func (m *MockDataServiceClient) AddData(ctx context.Context, in *datapb.AddDataRequest, opts ...grpc.CallOption) (*datapb.AddDataResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*datapb.AddDataResponse), args.Error(1)
}

func (m *MockDataServiceClient) GetData(ctx context.Context, in *datapb.GetDataRequest, opts ...grpc.CallOption) (*datapb.GetDataResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*datapb.GetDataResponse), args.Error(1)
}

func (m *MockDataServiceClient) UpdateData(ctx context.Context, in *datapb.UpdateDataRequest, opts ...grpc.CallOption) (*datapb.UpdateDataResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*datapb.UpdateDataResponse), args.Error(1)
}

func (m *MockDataServiceClient) DeleteData(ctx context.Context, in *datapb.DeleteDataRequest, opts ...grpc.CallOption) (*datapb.DeleteDataResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*datapb.DeleteDataResponse), args.Error(1)
}

func (m *MockDataServiceClient) ListData(ctx context.Context, in *datapb.ListDataRequest, opts ...grpc.CallOption) (*datapb.ListDataResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*datapb.ListDataResponse), args.Error(1)
}

func TestDataService_AddData(t *testing.T) {
	mockClient := new(MockDataServiceClient)
	mockLogger := new(mockLogger)

	dataService := &dataService{
		client: mockClient,
		logger: mockLogger,
	}

	ctx := context.Background()
	token := "test-token"
	dataItem := &datapb.DataItem{
		InfoType: "text",
		Info:     []byte("test info"),
		Meta:     "test meta",
	}

	ctxWithMetadata := metadata.AppendToOutgoingContext(ctx, "authorization", token)

	expectedRequest := &datapb.AddDataRequest{Data: dataItem}

	expectedResponse := &datapb.AddDataResponse{Id: 1}

	mockClient.On("AddData", ctxWithMetadata, expectedRequest).Return(expectedResponse, nil)

	id, err := dataService.AddData(ctx, token, dataItem)

	assert.NoError(t, err)
	assert.Equal(t, int32(1), id)
	mockClient.AssertExpectations(t)
}

func TestDataService_GetData(t *testing.T) {
	mockClient := new(MockDataServiceClient)
	mockLogger := new(mockLogger)

	dataService := &dataService{
		client: mockClient,
		logger: mockLogger,
	}

	ctx := context.Background()
	token := "test-token"
	dataID := int32(1)

	ctxWithMetadata := metadata.AppendToOutgoingContext(ctx, "authorization", token)

	expectedRequest := &datapb.GetDataRequest{Id: dataID}

	expectedData := &datapb.DataItem{
		Id:       dataID,
		InfoType: "text",
		Info:     []byte("test info"),
		Meta:     "test meta",
	}

	expectedResponse := &datapb.GetDataResponse{Data: expectedData}

	mockClient.On("GetData", ctxWithMetadata, expectedRequest).Return(expectedResponse, nil)

	dataItem, err := dataService.GetData(ctx, token, dataID)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, dataItem)
	mockClient.AssertExpectations(t)
}

func TestDataService_UpdateData(t *testing.T) {
	mockClient := new(MockDataServiceClient)
	mockLogger := new(mockLogger)

	dataService := &dataService{
		client: mockClient,
		logger: mockLogger,
	}

	ctx := context.Background()
	token := "test-token"
	dataItem := &datapb.DataItem{
		Id:       1,
		InfoType: "text",
		Info:     []byte("updated info"),
		Meta:     "updated meta",
	}

	ctxWithMetadata := metadata.AppendToOutgoingContext(ctx, "authorization", token)

	expectedRequest := &datapb.UpdateDataRequest{Data: dataItem}

	expectedResponse := &datapb.UpdateDataResponse{}

	mockClient.On("UpdateData", ctxWithMetadata, expectedRequest).Return(expectedResponse, nil)

	err := dataService.UpdateData(ctx, token, dataItem)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDataService_DeleteData(t *testing.T) {
	mockClient := new(MockDataServiceClient)
	mockLogger := new(mockLogger)

	dataService := &dataService{
		client: mockClient,
		logger: mockLogger,
	}

	ctx := context.Background()
	token := "test-token"
	dataID := int32(1)

	ctxWithMetadata := metadata.AppendToOutgoingContext(ctx, "authorization", token)

	expectedRequest := &datapb.DeleteDataRequest{Id: dataID}

	expectedResponse := &datapb.DeleteDataResponse{}

	mockClient.On("DeleteData", ctxWithMetadata, expectedRequest).Return(expectedResponse, nil)

	err := dataService.DeleteData(ctx, token, dataID)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDataService_AddData_Error(t *testing.T) {
	mockClient := new(MockDataServiceClient)
	mockLogger := new(mockLogger)

	dataService := &dataService{
		client: mockClient,
		logger: mockLogger,
	}

	ctx := context.Background()
	token := "test-token"
	dataItem := &datapb.DataItem{
		InfoType: "text",
		Info:     []byte("test info"),
		Meta:     "test meta",
	}

	ctxWithMetadata := metadata.AppendToOutgoingContext(ctx, "authorization", token)

	expectedRequest := &datapb.AddDataRequest{Data: dataItem}

	mockClient.On("AddData", ctxWithMetadata, expectedRequest).Return((*datapb.AddDataResponse)(nil), errors.New("test error"))

	id, err := dataService.AddData(ctx, token, dataItem)

	assert.Error(t, err)
	assert.Equal(t, int32(0), id)
	mockClient.AssertExpectations(t)
}

func TestDataService_ListData(t *testing.T) {
	mockClient := new(MockDataServiceClient)
	mockLogger := new(mockLogger)

	dataService := &dataService{
		client: mockClient,
		logger: mockLogger,
	}

	ctx := context.Background()
	token := "test-token"

	filter := &entity.DataFilter{
		InfoType: "text",
	}

	ctxWithMetadata := metadata.AppendToOutgoingContext(ctx, "authorization", token)

	expectedRequest := &datapb.ListDataRequest{
		InfoType: filter.InfoType,
	}

	expectedDataItems := []*datapb.DataItem{
		{
			Id:       1,
			InfoType: "text",
			Info:     []byte("test info 1"),
			Meta:     "test meta 1",
		},
		{
			Id:       2,
			InfoType: "text",
			Info:     []byte("test info 2"),
			Meta:     "test meta 2",
		},
	}

	expectedResponse := &datapb.ListDataResponse{
		DataItems: expectedDataItems,
	}

	mockClient.On("ListData", ctxWithMetadata, expectedRequest).Return(expectedResponse, nil)

	dataItems, err := dataService.ListData(ctx, token, filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedDataItems, dataItems)
	mockClient.AssertExpectations(t)
}

func TestDataService_ListData_Error(t *testing.T) {
	mockClient := new(MockDataServiceClient)
	mockLogger := new(mockLogger)

	dataService := &dataService{
		client: mockClient,
		logger: mockLogger,
	}

	ctx := context.Background()
	token := "test-token"

	filter := &entity.DataFilter{
		InfoType: "text",
	}

	ctxWithMetadata := metadata.AppendToOutgoingContext(ctx, "authorization", token)

	expectedRequest := &datapb.ListDataRequest{
		InfoType: filter.InfoType,
	}

	mockClient.On("ListData", ctxWithMetadata, expectedRequest).Return((*datapb.ListDataResponse)(nil), errors.New("test error"))

	dataItems, err := dataService.ListData(ctx, token, filter)

	assert.Error(t, err)
	assert.Nil(t, dataItems)
	mockClient.AssertExpectations(t)
}

type MockGRPCClient struct {
	DataClient datapb.DataServiceClient
}

func TestNewDataService(t *testing.T) {
	mockDataClient := &MockDataServiceClient{}

	mockGRPCClient := &GRPCClient{
		DataClient: mockDataClient,
	}

	mockLogger := &mockLogger{}

	dataService := NewDataService(mockGRPCClient, mockLogger)

	assert.NotNil(t, dataService)

	assert.Equal(t, mockDataClient, dataService.client)

	assert.Equal(t, mockLogger, dataService.logger)
}
