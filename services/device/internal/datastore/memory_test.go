package datastore

import (
	"testing"

	"github.com/velvetriddles/mini-smart-home/services/device/internal/model"
)

func TestMemoryStore_GetDevice(t *testing.T) {
	store := NewMemoryStore()

	// Создаем тестовое устройство
	testDevice := model.NewDevice("Test Device", model.DeviceTypeLamp, "Test Model", "Test Room")
	if err := store.SaveDevice(testDevice); err != nil {
		t.Fatalf("Failed to save test device: %v", err)
	}

	tests := []struct {
		name      string
		deviceID  string
		wantError bool
	}{
		{
			name:      "Existing device",
			deviceID:  testDevice.ID,
			wantError: false,
		},
		{
			name:      "Non-existing device",
			deviceID:  "non-existing-id",
			wantError: true,
		},
		{
			name:      "Empty ID",
			deviceID:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device, err := store.GetDevice(tt.deviceID)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if device == nil {
					t.Error("Device is nil")
				} else if device.ID != testDevice.ID {
					t.Errorf("Expected device ID %s, got %s", testDevice.ID, device.ID)
				}
			}
		})
	}
}

func TestMemoryStore_ListDevices(t *testing.T) {
	store := NewMemoryStore()

	// Добавляем устройства разных типов
	lamp1 := model.NewDevice("Lamp 1", model.DeviceTypeLamp, "Model A", "Room 1")
	lamp1.Status.Online = true
	store.SaveDevice(lamp1)

	lamp2 := model.NewDevice("Lamp 2", model.DeviceTypeLamp, "Model B", "Room 2")
	lamp2.Status.Online = false
	store.SaveDevice(lamp2)

	thermo := model.NewDevice("Thermostat", model.DeviceTypeThermostat, "Model C", "Room 3")
	thermo.Status.Online = true
	store.SaveDevice(thermo)

	tests := []struct {
		name       string
		deviceType string
		onlineOnly bool
		wantCount  int
	}{
		{
			name:       "All devices",
			deviceType: "",
			onlineOnly: false,
			wantCount:  3,
		},
		{
			name:       "Only lamps",
			deviceType: model.DeviceTypeLamp,
			onlineOnly: false,
			wantCount:  2,
		},
		{
			name:       "Only thermostats",
			deviceType: model.DeviceTypeThermostat,
			onlineOnly: false,
			wantCount:  1,
		},
		{
			name:       "Only online devices",
			deviceType: "",
			onlineOnly: true,
			wantCount:  2,
		},
		{
			name:       "Only online lamps",
			deviceType: model.DeviceTypeLamp,
			onlineOnly: true,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, err := store.ListDevices(tt.deviceType, tt.onlineOnly)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(devices) != tt.wantCount {
				t.Errorf("Expected %d devices, got %d", tt.wantCount, len(devices))
			}
		})
	}
}

func TestMemoryStore_SaveAndDeleteDevice(t *testing.T) {
	store := NewMemoryStore()

	// Создаем тестовое устройство
	testDevice := model.NewDevice("Test Device", model.DeviceTypeLamp, "Test Model", "Test Room")

	// Сохраняем устройство
	if err := store.SaveDevice(testDevice); err != nil {
		t.Fatalf("Failed to save device: %v", err)
	}

	// Проверяем, что устройство сохранено
	device, err := store.GetDevice(testDevice.ID)
	if err != nil {
		t.Fatalf("Failed to get device after save: %v", err)
	}
	if device.ID != testDevice.ID {
		t.Errorf("Expected device ID %s, got %s", testDevice.ID, device.ID)
	}

	// Удаляем устройство
	if err := store.DeleteDevice(testDevice.ID); err != nil {
		t.Fatalf("Failed to delete device: %v", err)
	}

	// Проверяем, что устройство удалено
	_, err = store.GetDevice(testDevice.ID)
	if err != ErrDeviceNotFound {
		t.Errorf("Expected ErrDeviceNotFound, got %v", err)
	}
}

func TestMemoryStore_UpdateDeviceStatus(t *testing.T) {
	store := NewMemoryStore()

	// Создаем тестовое устройство
	testDevice := model.NewDevice("Test Device", model.DeviceTypeLamp, "Test Model", "Test Room")
	testDevice.Status.Online = true
	store.SaveDevice(testDevice)

	// Создаем новый статус
	newStatus := model.NewDeviceStatus()
	newStatus.Online = false
	newStatus.BatteryLevel = 50

	// Обновляем статус
	if err := store.UpdateDeviceStatus(testDevice.ID, newStatus); err != nil {
		t.Fatalf("Failed to update device status: %v", err)
	}

	// Получаем устройство и проверяем статус
	device, err := store.GetDevice(testDevice.ID)
	if err != nil {
		t.Fatalf("Failed to get device: %v", err)
	}

	if device.Status.Online != false {
		t.Errorf("Expected online status false, got %v", device.Status.Online)
	}

	if device.Status.BatteryLevel != 50 {
		t.Errorf("Expected battery level 50, got %d", device.Status.BatteryLevel)
	}
}

func TestMemoryStore_UpdateDeviceParameter(t *testing.T) {
	store := NewMemoryStore()

	// Создаем тестовое устройство
	testDevice := model.NewDevice("Test Device", model.DeviceTypeLamp, "Test Model", "Test Room")
	testDevice.Status.Parameters[model.ParamPower] = "off"
	store.SaveDevice(testDevice)

	// Обновляем параметр
	if err := store.UpdateDeviceParameter(testDevice.ID, model.ParamPower, "on"); err != nil {
		t.Fatalf("Failed to update device parameter: %v", err)
	}

	// Получаем устройство и проверяем параметр
	device, err := store.GetDevice(testDevice.ID)
	if err != nil {
		t.Fatalf("Failed to get device: %v", err)
	}

	if value, ok := device.Status.Parameters[model.ParamPower]; !ok || value != "on" {
		t.Errorf("Expected parameter %s to be 'on', got %s", model.ParamPower, value)
	}
}

func TestMemoryStore_GetAllDevices(t *testing.T) {
	store := NewMemoryStore()

	// Добавляем несколько устройств
	for i := 0; i < 5; i++ {
		device := model.NewDevice("Test Device", model.DeviceTypeLamp, "Test Model", "Test Room")
		store.SaveDevice(device)
	}

	// Получаем все устройства
	devices := store.GetAllDevices()

	// Проверяем количество устройств
	if len(devices) != 5 {
		t.Errorf("Expected 5 devices, got %d", len(devices))
	}
}
