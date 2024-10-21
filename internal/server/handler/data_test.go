package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/internal/contextkey"
	"github.com/NikolosHGW/goph-keeper/internal/server/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockLogger struct{}

func (l *mockLogger) LogInfo(message string, err error) {}

type mockDataService struct {
	AddDataFunc     func(ctx context.Context, userID int, data *entity.UserData) (int, error)
	GetDataByIDFunc func(ctx context.Context, userID, dataID int) (*entity.UserData, error)
	UpdateDataFunc  func(ctx context.Context, userID int, data *entity.UserData) error
	DeleteDataFunc  func(ctx context.Context, userID, dataID int) error
}

func (m *mockDataService) AddData(ctx context.Context, userID int, data *entity.UserData) (int, error) {
	return m.AddDataFunc(ctx, userID, data)
}

func (m *mockDataService) GetDataByID(ctx context.Context, userID, dataID int) (*entity.UserData, error) {
	return m.GetDataByIDFunc(ctx, userID, dataID)
}

func (m *mockDataService) UpdateData(ctx context.Context, userID int, data *entity.UserData) error {
	return m.UpdateDataFunc(ctx, userID, data)
}

func (m *mockDataService) DeleteData(ctx context.Context, userID, dataID int) error {
	return m.DeleteDataFunc(ctx, userID, dataID)
}

func contextWithUserID(userID int) context.Context {
	return context.WithValue(context.Background(), contextkey.UserIDKey, userID)
}

func TestAddData(t *testing.T) {
	mockService := &mockDataService{}
	mockLogger := &mockLogger{}

	server := NewDataServer(mockService, mockLogger)

	tests := []struct {
		name          string
		ctx           context.Context
		request       *datapb.AddDataRequest
		setupMocks    func()
		expectedResp  *datapb.AddDataResponse
		expectedError error
	}{
		{
			name: "Success",
			ctx:  contextWithUserID(1),
			request: &datapb.AddDataRequest{
				Data: &datapb.DataItem{
					InfoType: "password",
					Info:     "mypassword",
					Meta:     "meta",
				},
			},
			setupMocks: func() {
				mockService.AddDataFunc = func(ctx context.Context, userID int, data *entity.UserData) (int, error) {
					if userID != 1 {
						t.Errorf("Expected userID 1, got %d", userID)
					}
					if data.InfoType != "password" || data.Info != "mypassword" || data.Meta != "meta" {
						t.Errorf("Unexpected data: %+v", data)
					}
					return 123, nil
				}
			},
			expectedResp:  &datapb.AddDataResponse{Id: 123},
			expectedError: nil,
		},
		{
			name:          "NoUserID",
			ctx:           context.Background(),
			request:       &datapb.AddDataRequest{},
			setupMocks:    func() {},
			expectedResp:  nil,
			expectedError: statusError(codes.Internal, "не удалось получить userID из контекста"),
		},
		{
			name: "DataServiceError",
			ctx:  contextWithUserID(1),
			request: &datapb.AddDataRequest{
				Data: &datapb.DataItem{
					InfoType: "password",
					Info:     "mypassword",
					Meta:     "meta",
				},
			},
			setupMocks: func() {
				mockService.AddDataFunc = func(ctx context.Context, userID int, data *entity.UserData) (int, error) {
					return 0, errors.New("database error")
				}
			},
			expectedResp:  nil,
			expectedError: statusError(codes.Internal, "ошибка при добавлении данных"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			resp, err := server.AddData(tt.ctx, tt.request)
			if !compareErrors(err, tt.expectedError) {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}
			if !compareAddDataResponse(resp, tt.expectedResp) {
				t.Errorf("Expected response: %v, got: %v", tt.expectedResp, resp)
			}
		})
	}
}

func TestGetData(t *testing.T) {
	mockService := &mockDataService{}
	mockLogger := &mockLogger{}

	server := NewDataServer(mockService, mockLogger)

	tests := []struct {
		name          string
		ctx           context.Context
		request       *datapb.GetDataRequest
		setupMocks    func()
		expectedResp  *datapb.GetDataResponse
		expectedError error
	}{
		{
			name: "Success",
			ctx:  contextWithUserID(1),
			request: &datapb.GetDataRequest{
				Id: 123,
			},
			setupMocks: func() {
				mockService.GetDataByIDFunc = func(ctx context.Context, userID, dataID int) (*entity.UserData, error) {
					if userID != 1 || dataID != 123 {
						t.Errorf("Unexpected userID or dataID: %d, %d", userID, dataID)
					}
					return &entity.UserData{
						ID:       123,
						UserID:   1,
						InfoType: "password",
						Info:     "mypassword",
						Meta:     "meta",
						Created:  time.Now(),
					}, nil
				}
			},
			expectedResp: &datapb.GetDataResponse{
				Data: &datapb.DataItem{
					Id:       123,
					InfoType: "password",
					Info:     "mypassword",
					Meta:     "meta",
				},
			},
			expectedError: nil,
		},
		{
			name:          "NoUserID",
			ctx:           context.Background(),
			request:       &datapb.GetDataRequest{},
			setupMocks:    func() {},
			expectedResp:  nil,
			expectedError: errors.New("userID не найден в контексте"),
		},
		{
			name: "DataServiceError",
			ctx:  contextWithUserID(1),
			request: &datapb.GetDataRequest{
				Id: 123,
			},
			setupMocks: func() {
				mockService.GetDataByIDFunc = func(ctx context.Context, userID, dataID int) (*entity.UserData, error) {
					return nil, errors.New("not found")
				}
			},
			expectedResp:  nil,
			expectedError: statusError(codes.NotFound, "данные не найдены"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			resp, err := server.GetData(tt.ctx, tt.request)
			if !compareErrors(err, tt.expectedError) {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}
			if tt.expectedResp != nil {
				if resp == nil || resp.Data == nil {
					t.Errorf("Expected response data, got nil")
				} else {
					if resp.Data.Id != tt.expectedResp.Data.Id ||
						resp.Data.InfoType != tt.expectedResp.Data.InfoType ||
						resp.Data.Info != tt.expectedResp.Data.Info ||
						resp.Data.Meta != tt.expectedResp.Data.Meta {
						t.Errorf("Response data does not match expected")
					}
				}
			} else if resp != nil {
				t.Errorf("Expected nil response, got: %v", resp)
			}
		})
	}
}

func TestUpdateData(t *testing.T) {
	mockService := &mockDataService{}
	mockLogger := &mockLogger{}

	server := NewDataServer(mockService, mockLogger)

	tests := []struct {
		name          string
		ctx           context.Context
		request       *datapb.UpdateDataRequest
		setupMocks    func()
		expectedResp  *datapb.UpdateDataResponse
		expectedError error
	}{
		{
			name: "Success",
			ctx:  contextWithUserID(1),
			request: &datapb.UpdateDataRequest{
				Data: &datapb.DataItem{
					Id:       123,
					InfoType: "password",
					Info:     "newpassword",
					Meta:     "newmeta",
					Created:  timestamppb.New(time.Now()),
				},
			},
			setupMocks: func() {
				mockService.UpdateDataFunc = func(ctx context.Context, userID int, data *entity.UserData) error {
					if userID != 1 || data.ID != 123 || data.Info != "newpassword" {
						t.Errorf("Unexpected data in UpdateData")
					}
					return nil
				}
			},
			expectedResp:  &datapb.UpdateDataResponse{},
			expectedError: nil,
		},
		{
			name:          "NoUserID",
			ctx:           context.Background(),
			request:       &datapb.UpdateDataRequest{},
			setupMocks:    func() {},
			expectedResp:  nil,
			expectedError: statusError(codes.Internal, "не удалось получить userID из контекста"),
		},
		{
			name: "DataServiceError",
			ctx:  contextWithUserID(1),
			request: &datapb.UpdateDataRequest{
				Data: &datapb.DataItem{
					Id:       123,
					InfoType: "password",
					Info:     "newpassword",
					Meta:     "newmeta",
					Created:  timestamppb.New(time.Now()),
				},
			},
			setupMocks: func() {
				mockService.UpdateDataFunc = func(ctx context.Context, userID int, data *entity.UserData) error {
					return errors.New("update failed")
				}
			},
			expectedResp:  nil,
			expectedError: statusError(codes.Internal, "ошибка при обновлении данных"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			resp, err := server.UpdateData(tt.ctx, tt.request)

			if !compareErrors(err, tt.expectedError) {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}

			if !compareUpdateDataResponse(resp, tt.expectedResp) {
				t.Errorf("Expected response: %v, got: %v", tt.expectedResp, resp)
			}
		})
	}
}

func TestDeleteData(t *testing.T) {
	mockService := &mockDataService{}
	mockLogger := &mockLogger{}

	server := NewDataServer(mockService, mockLogger)

	tests := []struct {
		name          string
		ctx           context.Context
		request       *datapb.DeleteDataRequest
		setupMocks    func()
		expectedResp  *datapb.DeleteDataResponse
		expectedError error
	}{
		{
			name: "Success",
			ctx:  contextWithUserID(1),
			request: &datapb.DeleteDataRequest{
				Id: 123,
			},
			setupMocks: func() {
				mockService.DeleteDataFunc = func(ctx context.Context, userID, dataID int) error {
					if userID != 1 || dataID != 123 {
						t.Errorf("Unexpected userID or dataID in DeleteData")
					}
					return nil
				}
			},
			expectedResp:  &datapb.DeleteDataResponse{},
			expectedError: nil,
		},
		{
			name:          "NoUserID",
			ctx:           context.Background(),
			request:       &datapb.DeleteDataRequest{},
			setupMocks:    func() {},
			expectedResp:  nil,
			expectedError: statusError(codes.Internal, "не удалось получить userID из контекста"),
		},
		{
			name: "DataServiceError",
			ctx:  contextWithUserID(1),
			request: &datapb.DeleteDataRequest{
				Id: 123,
			},
			setupMocks: func() {
				mockService.DeleteDataFunc = func(ctx context.Context, userID, dataID int) error {
					return errors.New("delete failed")
				}
			},
			expectedResp:  nil,
			expectedError: statusError(codes.Internal, "ошибка при удалении данных"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			resp, err := server.DeleteData(tt.ctx, tt.request)

			if !compareErrors(err, tt.expectedError) {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}

			if !compareDeleteDataResponse(resp, tt.expectedResp) {
				t.Errorf("Expected response: %v, got: %v", tt.expectedResp, resp)
			}
		})
	}
}

func compareErrors(got, want error) bool {
	if got == nil && want == nil {
		return true
	}
	if got == nil || want == nil {
		return false
	}
	return got.Error() == want.Error()
}

func statusError(code codes.Code, msg string) error {
	return status.Error(code, msg)
}

func compareAddDataResponse(got, want *datapb.AddDataResponse) bool {
	if got == nil && want == nil {
		return true
	}
	if got == nil || want == nil {
		return false
	}
	return got.Id == want.Id
}

func compareUpdateDataResponse(got, want *datapb.UpdateDataResponse) bool {
	if got == nil && want == nil {
		return true
	}
	if got == nil || want == nil {
		return false
	}
	return true
}

func compareDeleteDataResponse(got, want *datapb.DeleteDataResponse) bool {
	if got == nil && want == nil {
		return true
	}
	if got == nil || want == nil {
		return false
	}
	return true
}
