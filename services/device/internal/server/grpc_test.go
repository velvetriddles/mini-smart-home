package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velvetriddles/mini-smart-home/services/device/internal/datastore"
	"github.com/velvetriddles/mini-smart-home/services/device/internal/model"
	pb "github.com/velvetriddles/mini-smart-home/services/device/proto/smarthome/v1"
)

// MockDeviceStore реализует интерфейс DeviceStore для тестирования
type MockDeviceStore struct {
	mock.Mock
}

func (m *MockDeviceStore) GetDevice(id string) (*model.Device, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Device), args.Error(1)
}

func (m *MockDeviceStore) ListDevices(deviceType string, onlineOnly bool) ([]*model.Device, error) {
	args := m.Called(deviceType, onlineOnly)
	return args.Get(0).([]*model.Device), args.Error(1)
}

func (m *MockDeviceStore) SaveDevice(device *model.Device) error {
	args := m.Called(device)
	return args.Error(0)
}

func (m *MockDeviceStore) DeleteDevice(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDeviceStore) UpdateDeviceStatus(id string, status *model.DeviceStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockDeviceStore) UpdateDeviceParameter(id, key, value string) error {
	args := m.Called(id, key, value)
	return args.Error(0)
}

func (m *MockDeviceStore) GetAllDevices() []*model.Device {
	args := m.Called()
	return args.Get(0).([]*model.Device)
}

func (m *MockDeviceStore) AddTestDevices() {
	m.Called()
}

func TestControlDevice_TurnOn(t *testing.T) {
	// Создаем мок хранилища
	mockStore := new(MockDeviceStore)

	// Создаем тестовое устройство
	testDevice := model.NewDevice("Test Lamp", model.DeviceTypeLamp, "Test Model", "Room")
	testDevice.Status.Online = true
	testDevice.Status.Parameters[model.ParamPower] = "off"

	// Настраиваем ожидания мока
	mockStore.On("GetDevice", "test-id").Return(testDevice, nil)
	mockStore.On("SaveDevice", mock.AnythingOfType("*model.Device")).Return(nil)

	// Создаем сервер с мок-хранилищем
	server := NewGRPCServer(mockStore)

	// Создаем запрос на включение устройства
	req := &pb.ControlDeviceRequest{
		Id: "test-id",
		Command: &pb.Command{
			Action: model.CommandTurnOn,
		},
	}

	// Вызываем метод управления устройством
	resp, err := server.ControlDevice(context.Background(), req)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Command executed successfully", resp.Status)

	// Проверяем, что устройство было обновлено в хранилище
	mockStore.AssertCalled(t, "SaveDevice", mock.AnythingOfType("*model.Device"))

	// Проверяем, что в первый аргумент SaveDevice передано устройство
	// с обновленным параметром power = "on"
	savedDevice := mockStore.Calls[1].Arguments.Get(0).(*model.Device)
	assert.Equal(t, "on", savedDevice.Status.Parameters[model.ParamPower])
}

func TestControlDevice_OfflineDevice(t *testing.T) {
	// Создаем мок хранилища
	mockStore := new(MockDeviceStore)

	// Создаем тестовое устройство (оффлайн)
	testDevice := model.NewDevice("Test Lamp", model.DeviceTypeLamp, "Test Model", "Room")
	testDevice.Status.Online = false

	// Настраиваем ожидания мока
	mockStore.On("GetDevice", "test-id").Return(testDevice, nil)

	// Создаем сервер с мок-хранилищем
	server := NewGRPCServer(mockStore)

	// Создаем запрос на включение устройства
	req := &pb.ControlDeviceRequest{
		Id: "test-id",
		Command: &pb.Command{
			Action: model.CommandTurnOn,
		},
	}

	// Вызываем метод управления устройством
	resp, err := server.ControlDevice(context.Background(), req)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "Device is offline", resp.Status)
	assert.Equal(t, "Cannot control offline device", resp.Error)

	// Проверяем, что SaveDevice НЕ вызывался
	mockStore.AssertNotCalled(t, "SaveDevice", mock.Anything)
}

func TestControlDevice_DeviceNotFound(t *testing.T) {
	// Создаем мок хранилища
	mockStore := new(MockDeviceStore)

	// Настраиваем ожидания мока - устройство не найдено
	mockStore.On("GetDevice", "non-existent-id").Return(nil, datastore.ErrDeviceNotFound)

	// Создаем сервер с мок-хранилищем
	server := NewGRPCServer(mockStore)

	// Создаем запрос
	req := &pb.ControlDeviceRequest{
		Id: "non-existent-id",
		Command: &pb.Command{
			Action: model.CommandTurnOn,
		},
	}

	// Вызываем метод управления устройством
	resp, err := server.ControlDevice(context.Background(), req)

	// Проверяем результаты - должна быть ошибка
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Проверяем, что SaveDevice НЕ вызывался
	mockStore.AssertNotCalled(t, "SaveDevice", mock.Anything)
}

func TestControlDevice_InvalidAction(t *testing.T) {
	// Создаем мок хранилища
	mockStore := new(MockDeviceStore)

	// Создаем тестовое устройство
	testDevice := model.NewDevice("Test Lamp", model.DeviceTypeLamp, "Test Model", "Room")
	testDevice.Status.Online = true

	// Настраиваем ожидания мока
	mockStore.On("GetDevice", "test-id").Return(testDevice, nil)

	// Создаем сервер с мок-хранилищем
	server := NewGRPCServer(mockStore)

	// Создаем запрос с неподдерживаемым действием
	req := &pb.ControlDeviceRequest{
		Id: "test-id",
		Command: &pb.Command{
			Action: "invalid_action",
		},
	}

	// Вызываем метод управления устройством
	resp, err := server.ControlDevice(context.Background(), req)

	// Проверяем результаты
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "Unsupported action", resp.Status)

	// Проверяем, что SaveDevice НЕ вызывался
	mockStore.AssertNotCalled(t, "SaveDevice", mock.Anything)
}
