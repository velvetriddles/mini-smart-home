package datastore

import (
	"errors"
	"sync"
	"time"

	"github.com/velvetriddles/mini-smart-home/services/device/internal/model"
)

var (
	// ErrDeviceNotFound означает, что устройство не найдено в хранилище
	ErrDeviceNotFound = errors.New("device not found")

	// ErrInvalidDeviceID означает, что указан неверный идентификатор устройства
	ErrInvalidDeviceID = errors.New("invalid device ID")
)

// DeviceStore представляет интерфейс хранилища устройств
type DeviceStore interface {
	// GetDevice возвращает устройство по ID
	GetDevice(id string) (*model.Device, error)

	// ListDevices возвращает список устройств, опционально фильтрованных
	ListDevices(deviceType string, onlineOnly bool) ([]*model.Device, error)

	// SaveDevice сохраняет устройство в хранилище
	SaveDevice(device *model.Device) error

	// DeleteDevice удаляет устройство из хранилища
	DeleteDevice(id string) error

	// UpdateDeviceStatus обновляет статус устройства
	UpdateDeviceStatus(id string, status *model.DeviceStatus) error

	// UpdateDeviceParameter обновляет параметр устройства
	UpdateDeviceParameter(id, key, value string) error

	// GetAllDevices возвращает все устройства для стриминга статусов
	GetAllDevices() []*model.Device

	// AddTestDevices добавляет тестовые устройства для разработки
	AddTestDevices()
}

// MemoryStore реализует хранилище устройств в памяти
type MemoryStore struct {
	devices map[string]*model.Device
	mu      sync.RWMutex
}

// NewMemoryStore создает новый экземпляр in-memory хранилища
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		devices: make(map[string]*model.Device),
	}
}

// GetDevice возвращает устройство по ID
func (s *MemoryStore) GetDevice(id string) (*model.Device, error) {
	if id == "" {
		return nil, ErrInvalidDeviceID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	device, ok := s.devices[id]
	if !ok {
		return nil, ErrDeviceNotFound
	}

	return device, nil
}

// ListDevices возвращает список устройств, опционально фильтрованных
func (s *MemoryStore) ListDevices(deviceType string, onlineOnly bool) ([]*model.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Device

	for _, device := range s.devices {
		// Фильтрация по типу устройства, если указан
		if deviceType != "" && device.Type != deviceType {
			continue
		}

		// Фильтрация по онлайн-статусу, если указано
		if onlineOnly && !device.Status.Online {
			continue
		}

		result = append(result, device)
	}

	return result, nil
}

// SaveDevice сохраняет устройство в хранилище
func (s *MemoryStore) SaveDevice(device *model.Device) error {
	if device.ID == "" {
		return ErrInvalidDeviceID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.devices[device.ID] = device
	return nil
}

// DeleteDevice удаляет устройство из хранилища
func (s *MemoryStore) DeleteDevice(id string) error {
	if id == "" {
		return ErrInvalidDeviceID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.devices[id]; !ok {
		return ErrDeviceNotFound
	}

	delete(s.devices, id)
	return nil
}

// UpdateDeviceStatus обновляет статус устройства
func (s *MemoryStore) UpdateDeviceStatus(id string, status *model.DeviceStatus) error {
	if id == "" {
		return ErrInvalidDeviceID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	device, ok := s.devices[id]
	if !ok {
		return ErrDeviceNotFound
	}

	device.Status = status
	device.LastUpdated = time.Now()

	return nil
}

// UpdateDeviceParameter обновляет параметр устройства
func (s *MemoryStore) UpdateDeviceParameter(id, key, value string) error {
	if id == "" || key == "" {
		return ErrInvalidDeviceID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	device, ok := s.devices[id]
	if !ok {
		return ErrDeviceNotFound
	}

	if device.Status == nil {
		device.Status = model.NewDeviceStatus()
	}

	device.Status.Parameters[key] = value
	device.LastUpdated = time.Now()

	return nil
}

// GetAllDevices возвращает все устройства для стриминга статусов
func (s *MemoryStore) GetAllDevices() []*model.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]*model.Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}

	return devices
}

// AddTestDevices добавляет тестовые устройства для разработки
func (s *MemoryStore) AddTestDevices() {
	// Лампа в гостиной
	lamp := model.NewDevice("Лампа гостиная", model.DeviceTypeLamp, "Philips Hue", "Гостиная")
	lamp.Status.Parameters[model.ParamPower] = "on"
	lamp.Status.Parameters[model.ParamLevel] = "80"
	lamp.Status.Parameters[model.ParamColor] = "#FFFFFF"
	lamp.Tags = []string{"освещение", "гостиная"}
	s.SaveDevice(lamp)

	// Розетка на кухне
	socket := model.NewDevice("Розетка кухня", model.DeviceTypeSocket, "TP-Link Kasa", "Кухня")
	socket.Status.Parameters[model.ParamPower] = "off"
	socket.Tags = []string{"розетка", "кухня"}
	s.SaveDevice(socket)

	// Термостат в спальне
	thermostat := model.NewDevice("Термостат спальня", model.DeviceTypeThermostat, "Nest", "Спальня")
	thermostat.Status.Parameters[model.ParamPower] = "on"
	thermostat.Status.Parameters[model.ParamTemperature] = "22.5"
	thermostat.Status.Parameters[model.ParamMode] = "auto"
	thermostat.Tags = []string{"климат", "спальня"}
	s.SaveDevice(thermostat)

	// Датчик движения в коридоре
	sensor := model.NewDevice("Датчик движения коридор", model.DeviceTypeSensor, "Aqara", "Коридор")
	sensor.Status.Parameters[model.ParamLastMotion] = "1714580400" // Unix timestamp
	sensor.Status.BatteryLevel = 85
	sensor.Tags = []string{"датчик", "безопасность", "коридор"}
	s.SaveDevice(sensor)

	// Камера у входной двери
	camera := model.NewDevice("Камера входная дверь", model.DeviceTypeCamera, "Hikvision", "Прихожая")
	camera.Status.Parameters[model.ParamPower] = "on"
	camera.Status.Parameters[model.ParamMode] = "motion"
	camera.Tags = []string{"безопасность", "видеонаблюдение"}
	s.SaveDevice(camera)
}
