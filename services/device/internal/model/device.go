package model

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/velvetriddles/mini-smart-home/proto/smarthome/v1"
)

// Константы для типов устройств
const (
	DeviceTypeLamp       = "lamp"
	DeviceTypeSocket     = "socket"
	DeviceTypeThermostat = "thermostat"
	DeviceTypeSensor     = "sensor"
	DeviceTypeSwitch     = "switch"
	DeviceTypeCamera     = "camera"
)

// Константы для команд
const (
	CommandTurnOn    = "turn_on"
	CommandTurnOff   = "turn_off"
	CommandSetLevel  = "set_level"
	CommandSetTemp   = "set_temperature"
	CommandSetColor  = "set_color"
	CommandSetMode   = "set_mode"
	CommandRefresh   = "refresh"
	CommandReboot    = "reboot"
	CommandCalibrate = "calibrate"
)

// Константы для параметров статуса
const (
	ParamPower       = "power"
	ParamLevel       = "level"
	ParamTemperature = "temperature"
	ParamHumidity    = "humidity"
	ParamColor       = "color"
	ParamMode        = "mode"
	ParamLastMotion  = "last_motion"
)

// Device представляет устройство умного дома в системе
type Device struct {
	ID          string        // Уникальный идентификатор устройства
	Name        string        // Имя устройства
	Type        string        // Тип устройства
	Model       string        // Модель устройства
	Room        string        // Комната/расположение
	Status      *DeviceStatus // Текущий статус устройства
	Tags        []string      // Теги устройства
	LastUpdated time.Time     // Время последнего обновления
}

// DeviceStatus представляет текущий статус устройства
type DeviceStatus struct {
	Online         bool              // Онлайн статус
	Parameters     map[string]string // Динамические параметры
	BatteryLevel   int32             // Уровень батареи (если применимо)
	SignalStrength float32           // Сила сигнала (если применимо)
}

// NewDevice создает новое устройство с уникальным ID
func NewDevice(name, deviceType, model, room string) *Device {
	return &Device{
		ID:          uuid.New().String(),
		Name:        name,
		Type:        deviceType,
		Model:       model,
		Room:        room,
		Status:      NewDeviceStatus(),
		Tags:        []string{},
		LastUpdated: time.Now(),
	}
}

// NewDeviceStatus создает новый статус устройства
func NewDeviceStatus() *DeviceStatus {
	return &DeviceStatus{
		Online:         true,
		Parameters:     make(map[string]string),
		BatteryLevel:   100,
		SignalStrength: 90.0,
	}
}

// ToProto конвертирует Device в protobuf-представление
func (d *Device) ToProto() *pb.Device {
	return &pb.Device{
		Id:     d.ID,
		Name:   d.Name,
		Type:   d.Type,
		Model:  d.Model,
		Room:   d.Room,
		Status: d.Status.ToProto(),
		Tags:   d.Tags,
	}
}

// ToProto конвертирует DeviceStatus в protobuf-представление
func (s *DeviceStatus) ToProto() *pb.DeviceStatus {
	return &pb.DeviceStatus{
		Online:         s.Online,
		Parameters:     s.Parameters,
		BatteryLevel:   s.BatteryLevel,
		SignalStrength: s.SignalStrength,
	}
}

// UpdateStatus обновляет статус устройства
func (d *Device) UpdateStatus(status *DeviceStatus) {
	d.Status = status
	d.LastUpdated = time.Now()
}

// UpdateParameterValue обновляет значение отдельного параметра
func (d *Device) UpdateParameterValue(key, value string) {
	if d.Status == nil {
		d.Status = NewDeviceStatus()
	}
	d.Status.Parameters[key] = value
	d.LastUpdated = time.Now()
}

// CreateTimestamp создает объект Timestamp из времени
func CreateTimestamp(t time.Time) *pb.Timestamp {
	return &pb.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// TimestampToTime конвертирует Timestamp в time.Time
func TimestampToTime(ts *pb.Timestamp) time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}
