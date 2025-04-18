package server

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/velvetriddles/mini-smart-home/proto/smarthome/v1"
	"github.com/velvetriddles/mini-smart-home/services/device/internal/datastore"
	"github.com/velvetriddles/mini-smart-home/services/device/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer реализует интерфейс gRPC сервера для Device Service
type GRPCServer struct {
	pb.UnimplementedDeviceServiceServer
	store datastore.DeviceStore
}

// NewGRPCServer создает новый экземпляр gRPC сервера
func NewGRPCServer(store datastore.DeviceStore) *GRPCServer {
	return &GRPCServer{
		store: store,
	}
}

// GetDevice реализует gRPC метод для получения устройства по ID
func (s *GRPCServer) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	log.Printf("GetDevice request for ID: %s", req.Id)

	device, err := s.store.GetDevice(req.Id)
	if err != nil {
		if err == datastore.ErrDeviceNotFound {
			return nil, status.Errorf(codes.NotFound, "device with ID %s not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to get device: %v", err)
	}

	return &pb.GetDeviceResponse{
		Device: device.ToProto(),
	}, nil
}

// ListDevices реализует gRPC метод для получения списка устройств с фильтрацией
func (s *GRPCServer) ListDevices(ctx context.Context, req *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	log.Printf("ListDevices request with type filter: %s, online only: %v", req.Type, req.OnlineOnly)

	devices, err := s.store.ListDevices(req.Type, req.OnlineOnly)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list devices: %v", err)
	}

	protoDevices := make([]*pb.Device, 0, len(devices))
	for _, device := range devices {
		protoDevices = append(protoDevices, device.ToProto())
	}

	return &pb.ListDevicesResponse{
		Devices:    protoDevices,
		TotalCount: int32(len(protoDevices)),
	}, nil
}

// ControlDevice реализует gRPC метод для управления устройством
func (s *GRPCServer) ControlDevice(ctx context.Context, req *pb.ControlDeviceRequest) (*pb.ControlDeviceResponse, error) {
	log.Printf("ControlDevice request for ID: %s, action: %s", req.Id, req.Command.Action)

	// Проверяем существование устройства
	device, err := s.store.GetDevice(req.Id)
	if err != nil {
		if err == datastore.ErrDeviceNotFound {
			return nil, status.Errorf(codes.NotFound, "device with ID %s not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to get device: %v", err)
	}

	// Проверяем, что устройство онлайн
	if !device.Status.Online {
		return &pb.ControlDeviceResponse{
			Success: false,
			Status:  "Device is offline",
			Error:   "Cannot control offline device",
		}, nil
	}

	// Обрабатываем команду в зависимости от типа действия
	switch req.Command.Action {
	case model.CommandTurnOn:
		device.UpdateParameterValue(model.ParamPower, "on")
	case model.CommandTurnOff:
		device.UpdateParameterValue(model.ParamPower, "off")
	case model.CommandSetLevel:
		if level, ok := req.Command.Parameters[model.ParamLevel]; ok {
			device.UpdateParameterValue(model.ParamLevel, level)
		} else {
			return &pb.ControlDeviceResponse{
				Success: false,
				Status:  "Missing parameter",
				Error:   "Level parameter is required for set_level action",
			}, nil
		}
	case model.CommandSetTemp:
		if temp, ok := req.Command.Parameters[model.ParamTemperature]; ok {
			device.UpdateParameterValue(model.ParamTemperature, temp)
		} else {
			return &pb.ControlDeviceResponse{
				Success: false,
				Status:  "Missing parameter",
				Error:   "Temperature parameter is required for set_temperature action",
			}, nil
		}
	case model.CommandSetColor:
		if color, ok := req.Command.Parameters[model.ParamColor]; ok {
			device.UpdateParameterValue(model.ParamColor, color)
		} else {
			return &pb.ControlDeviceResponse{
				Success: false,
				Status:  "Missing parameter",
				Error:   "Color parameter is required for set_color action",
			}, nil
		}
	case model.CommandSetMode:
		if mode, ok := req.Command.Parameters[model.ParamMode]; ok {
			device.UpdateParameterValue(model.ParamMode, mode)
		} else {
			return &pb.ControlDeviceResponse{
				Success: false,
				Status:  "Missing parameter",
				Error:   "Mode parameter is required for set_mode action",
			}, nil
		}
	default:
		return &pb.ControlDeviceResponse{
			Success: false,
			Status:  "Unsupported action",
			Error:   "Action not supported for this device type",
		}, nil
	}

	// Обновляем устройство в хранилище
	if err := s.store.SaveDevice(device); err != nil {
		return &pb.ControlDeviceResponse{
			Success: false,
			Status:  "Internal error",
			Error:   "Failed to update device state",
		}, nil
	}

	return &pb.ControlDeviceResponse{
		Success: true,
		Status:  "Command executed successfully",
	}, nil
}

// SendCommand реализует gRPC метод для отправки команды на устройство
func (s *GRPCServer) SendCommand(ctx context.Context, cmd *pb.Command) (*pb.CommandResult, error) {
	log.Printf("SendCommand request for device ID: %s, action: %s", cmd.DeviceId, cmd.Action)

	// Проверяем существование устройства
	device, err := s.store.GetDevice(cmd.DeviceId)
	if err != nil {
		if err == datastore.ErrDeviceNotFound {
			return nil, status.Errorf(codes.NotFound, "device with ID %s not found", cmd.DeviceId)
		}
		return nil, status.Errorf(codes.Internal, "failed to get device: %v", err)
	}

	// Обрабатываем команду аналогично ControlDevice
	// Здесь для примера просто обновляем устройство и возвращаем успешный результат
	// В реальном приложении здесь был бы код для отправки команды на физическое устройство через MQTT

	// Обновляем время команды, если не задано
	timeNow := time.Now()
	if cmd.Time == nil {
		cmd.Time = &pb.Timestamp{
			Seconds: timeNow.Unix(),
			Nanos:   int32(timeNow.Nanosecond()),
		}
	}

	// Создаем объект результата
	result := &pb.CommandResult{
		DeviceId:     cmd.DeviceId,
		Success:      true,
		Status:       "Command sent successfully",
		ResponseData: make(map[string]string),
		Time: &pb.Timestamp{
			Seconds: timeNow.Unix(),
			Nanos:   int32(timeNow.Nanosecond()),
		},
	}

	// Для демонстрации возвращаем текущие параметры устройства в ответе
	if device.Status != nil && device.Status.Parameters != nil {
		for k, v := range device.Status.Parameters {
			result.ResponseData[k] = v
		}
	}

	return result, nil
}

// StreamStatuses реализует gRPC метод для потоковой передачи статусов устройств
func (s *GRPCServer) StreamStatuses(stream pb.DeviceService_StreamStatusesServer) error {
	log.Println("StreamStatuses: Starting status stream")

	// Получаем первый запрос для инициализации подписки
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return status.Errorf(codes.Internal, "failed to receive initial request: %v", err)
	}

	log.Printf("StreamStatuses: Received subscription request for device ID: %s, subscribe all: %v",
		req.DeviceId, req.SubscribeAll)

	// Определяем список устройств для мониторинга
	var deviceIDs []string
	if req.SubscribeAll {
		// Мониторим все устройства
		devices := s.store.GetAllDevices()
		for _, device := range devices {
			deviceIDs = append(deviceIDs, device.ID)
		}
	} else if req.DeviceId != "" {
		// Мониторим только одно устройство
		deviceIDs = append(deviceIDs, req.DeviceId)
	} else {
		return status.Errorf(codes.InvalidArgument,
			"device_id must be specified or subscribe_all must be true")
	}

	// Для простоты демонстрации отправляем статусы раз в 5 секунд
	// В реальном приложении здесь была бы подписка на Kafka/MQTT
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Канал для получения сигнала о завершении стрима
	done := make(chan struct{})

	// Горутина для прослушивания новых запросов от клиента
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Printf("StreamStatuses: Error receiving request: %v", err)
				close(done)
				return
			}
		}
	}()

	// Основной цикл отправки статусов
	for {
		select {
		case <-done:
			log.Println("StreamStatuses: Client disconnected")
			return nil
		case <-ticker.C:
			// Отправляем текущий статус для каждого устройства в списке
			for _, id := range deviceIDs {
				device, err := s.store.GetDevice(id)
				if err != nil {
					log.Printf("StreamStatuses: Error getting device %s: %v", id, err)
					continue
				}

				now := time.Now()
				resp := &pb.StatusResponse{
					DeviceId: device.ID,
					Status:   device.Status.ToProto(),
					Time: &pb.Timestamp{
						Seconds: now.Unix(),
						Nanos:   int32(now.Nanosecond()),
					},
				}

				if err := stream.Send(resp); err != nil {
					log.Printf("StreamStatuses: Error sending status update: %v", err)
					return err
				}

				log.Printf("StreamStatuses: Status sent for device %s", device.ID)
			}
		}
	}
}
