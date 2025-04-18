package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	pb "github.com/velvetriddles/mini-smart-home/proto/smarthome/v1"
	"github.com/velvetriddles/mini-smart-home/services/voice/internal/model"
	"github.com/velvetriddles/mini-smart-home/services/voice/internal/nlu"
)

// MockStream - мок для VoiceService_RecognizeCommandServer
type MockStream struct {
	mock.Mock
	ctx context.Context
	grpc.ServerStream
}

func (m *MockStream) Context() context.Context {
	return m.ctx
}

func (m *MockStream) Send(resp *pb.VoiceResponse) error {
	args := m.Called(resp)
	return args.Error(0)
}

func (m *MockStream) Recv() (*pb.VoiceRequest, error) {
	args := m.Called()
	return args.Get(0).(*pb.VoiceRequest), args.Error(1)
}

// MockDeviceClient - мок для DeviceServiceClient
type MockDeviceClient struct {
	mock.Mock
}

func (m *MockDeviceClient) GetDevice(ctx context.Context, in *pb.GetDeviceRequest, opts ...grpc.CallOption) (*pb.GetDeviceResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.GetDeviceResponse), args.Error(1)
}

func (m *MockDeviceClient) ListDevices(ctx context.Context, in *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ListDevicesResponse), args.Error(1)
}

func (m *MockDeviceClient) ControlDevice(ctx context.Context, in *pb.ControlDeviceRequest, opts ...grpc.CallOption) (*pb.ControlDeviceResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ControlDeviceResponse), args.Error(1)
}

func (m *MockDeviceClient) SendCommand(ctx context.Context, in *pb.Command, opts ...grpc.CallOption) (*pb.CommandResult, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.CommandResult), args.Error(1)
}

func (m *MockDeviceClient) StreamStatuses(ctx context.Context, opts ...grpc.CallOption) (pb.DeviceService_StreamStatusesClient, error) {
	args := m.Called(ctx)
	return args.Get(0).(pb.DeviceService_StreamStatusesClient), args.Error(1)
}

// MockAuthClient - мок для AuthServiceClient
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.LoginResponse), args.Error(1)
}

func (m *MockAuthClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.LogoutResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.LogoutResponse), args.Error(1)
}

func (m *MockAuthClient) Refresh(ctx context.Context, in *pb.RefreshRequest, opts ...grpc.CallOption) (*pb.RefreshResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.RefreshResponse), args.Error(1)
}

func (m *MockAuthClient) ValidateToken(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateTokenResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ValidateTokenResponse), args.Error(1)
}

func TestGRPCServer_handleTextCommand(t *testing.T) {
	// Создаем моки
	mockDeviceClient := new(MockDeviceClient)
	mockAuthClient := new(MockAuthClient)

	// Создаем сервер с моками
	server := &GRPCServer{
		nluEngine:     nlu.NewSimpleNLU(),
		deviceClient:  mockDeviceClient,
		authClient:    mockAuthClient,
		sessionTokens: make(map[string]string),
	}

	// Настраиваем ожидаемый ответ от DeviceService при запросе списка устройств
	mockDeviceClient.On("ListDevices", mock.Anything, &pb.ListDevicesRequest{
		Type:       "лампа",
		OnlineOnly: true,
	}).Return(&pb.ListDevicesResponse{
		Devices: []*pb.Device{
			{
				Id:   "lamp-123",
				Name: "Настольная лампа",
				Type: "лампа",
				Room: "гостиная",
				Status: &pb.DeviceStatus{
					Online: true,
				},
			},
		},
	}, nil)

	// Настраиваем ожидаемый ответ от DeviceService при отправке команды
	mockDeviceClient.On("ControlDevice", mock.Anything, &pb.ControlDeviceRequest{
		Id: "lamp-123",
		Command: &pb.Command{
			DeviceId:   "lamp-123",
			Action:     "turn_on",
			Parameters: make(map[string]string),
		},
	}).Return(&pb.ControlDeviceResponse{
		Success: true,
		Status:  "OK",
	}, nil)

	// Тестируем обработку команды "включи лампу в гостиной"
	sessionID := "test-session-123"
	response := server.handleTextCommand(context.Background(), "включи лампу в гостиной", sessionID)

	// Проверяем результаты
	assert.Equal(t, sessionID, response.SessionId)
	assert.True(t, response.Successful)
	assert.Equal(t, model.IntentTurnOn, response.RecognizedIntent.Name)
	assert.Contains(t, response.ResponseText, "успешно включено")

	// Проверяем, что были вызваны нужные методы с нужными параметрами
	mockDeviceClient.AssertExpectations(t)
}

func TestHandleTurnOnIntent(t *testing.T) {
	// Создаем intent для тестирования
	intent := model.NewIntent(model.IntentTurnOn, 0.8)
	intent.AddEntity(model.EntityDevice, "лампа", "лампу", 0.9)
	intent.AddEntity(model.EntityRoom, "гостиная", "гостиной", 0.9)
	intent.AddParameter("action", "turn_on")

	// Создаем моки
	mockDeviceClient := new(MockDeviceClient)
	mockAuthClient := new(MockAuthClient)

	// Создаем сервер с моками
	server := &GRPCServer{
		nluEngine:     nlu.NewSimpleNLU(),
		deviceClient:  mockDeviceClient,
		authClient:    mockAuthClient,
		sessionTokens: make(map[string]string),
	}

	// Настраиваем ожидаемый ответ от DeviceService при запросе списка устройств
	mockDeviceClient.On("ListDevices", mock.Anything, mock.Anything).Return(&pb.ListDevicesResponse{
		Devices: []*pb.Device{
			{
				Id:   "lamp-123",
				Name: "Настольная лампа",
				Type: "лампа",
				Room: "гостиная",
				Status: &pb.DeviceStatus{
					Online: true,
				},
			},
		},
	}, nil)

	// Настраиваем ожидаемый ответ от DeviceService при отправке команды
	mockDeviceClient.On("ControlDevice", mock.Anything, mock.Anything).Return(&pb.ControlDeviceResponse{
		Success: true,
		Status:  "OK",
	}, nil)

	// Вызываем метод
	responseText, successful, errMsg := server.handleDeviceControlIntent(context.Background(), intent, "test-session")

	// Проверяем результаты
	assert.True(t, successful)
	assert.Empty(t, errMsg)
	assert.Contains(t, responseText, "успешно включено")

	// Проверяем, что были вызваны нужные методы
	mockDeviceClient.AssertCalled(t, "ListDevices", mock.Anything, mock.Anything)
	mockDeviceClient.AssertCalled(t, "ControlDevice", mock.Anything, mock.Anything)
}
