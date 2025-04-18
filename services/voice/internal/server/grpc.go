package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	pb "github.com/velvetriddles/mini-smart-home/proto/smarthome/v1"
	"github.com/velvetriddles/mini-smart-home/services/voice/internal/model"
	"github.com/velvetriddles/mini-smart-home/services/voice/internal/nlu"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GRPCServer представляет реализацию сервера VoiceService
type GRPCServer struct {
	pb.UnimplementedVoiceServiceServer
	nluEngine     *nlu.SimpleNLU
	deviceClient  pb.DeviceServiceClient
	authClient    pb.AuthServiceClient
	sessionTokens map[string]string
}

// NewGRPCServer создает новый экземпляр gRPC-сервера для голосового управления
func NewGRPCServer(deviceConn *grpc.ClientConn, authConn *grpc.ClientConn) *GRPCServer {
	return &GRPCServer{
		nluEngine:     nlu.NewSimpleNLU(),
		deviceClient:  pb.NewDeviceServiceClient(deviceConn),
		authClient:    pb.NewAuthServiceClient(authConn),
		sessionTokens: make(map[string]string),
	}
}

// RecognizeCommand обрабатывает bi-directional поток голосовых команд
func (s *GRPCServer) RecognizeCommand(stream pb.VoiceService_RecognizeCommandServer) error {
	// Получаем метаданные запроса для извлечения токена авторизации
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "метаданные не найдены")
	}

	// Извлекаем токен из метаданных
	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return status.Error(codes.Unauthenticated, "токен авторизации не найден")
	}

	// Проверяем формат токена (Bearer token)
	tokenParts := strings.Split(authHeader[0], " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		return status.Error(codes.Unauthenticated, "неверный формат токена")
	}
	token := tokenParts[1]

	// Проверяем валидность токена через AuthService
	validateResp, err := s.authClient.ValidateToken(stream.Context(), &pb.ValidateTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		log.Printf("Ошибка при проверке токена: %v", err)
		return status.Error(codes.Unauthenticated, "ошибка проверки токена")
	}
	if !validateResp.Valid {
		return status.Error(codes.Unauthenticated, "недействительный токен")
	}

	// Основной цикл обработки потока
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("Ошибка при получении команды: %v", err)
			return status.Error(codes.Internal, "ошибка получения команды")
		}

		// Сохраняем токен для текущей сессии
		if req.SessionId != "" {
			s.sessionTokens[req.SessionId] = token
		}

		// Обрабатываем команду
		var response *pb.VoiceResponse
		switch req.Input.(type) {
		case *pb.VoiceRequest_Text:
			// Обрабатываем текстовый ввод
			textInput := req.GetText()
			response = s.handleTextCommand(stream.Context(), textInput, req.SessionId)
		case *pb.VoiceRequest_AudioData:
			// Аудио-обработка пока не реализована, возвращаем ошибку
			response = &pb.VoiceResponse{
				SessionId:    req.SessionId,
				ResponseText: "Обработка аудио-данных пока не поддерживается.",
				Successful:   false,
				Error:        "audio_not_supported",
			}
		default:
			response = &pb.VoiceResponse{
				SessionId:    req.SessionId,
				ResponseText: "Неизвестный тип входных данных.",
				Successful:   false,
				Error:        "invalid_input_type",
			}
		}

		// Отправляем ответ
		if err := stream.Send(response); err != nil {
			log.Printf("Ошибка при отправке ответа: %v", err)
			return status.Error(codes.Internal, "ошибка отправки ответа")
		}
	}
}

// ListIntents возвращает список поддерживаемых интентов
func (s *GRPCServer) ListIntents(ctx context.Context, _ *emptypb.Empty) (*pb.IntentList, error) {
	supportedIntents := s.nluEngine.GetSupportedIntents()
	result := &pb.IntentList{
		Intents: make([]*pb.IntentDefinition, 0, len(supportedIntents)),
	}

	for _, intent := range supportedIntents {
		pbIntent := &pb.IntentDefinition{
			Name:        intent.Name,
			Description: intent.Description,
			Examples:    intent.Examples,
			Entities:    make([]*pb.EntityDefinition, 0, len(intent.Entities)),
		}

		for _, entity := range intent.Entities {
			pbIntent.Entities = append(pbIntent.Entities, &pb.EntityDefinition{
				Name:        entity.Name,
				Description: entity.Description,
			})
		}

		result.Intents = append(result.Intents, pbIntent)
	}

	return result, nil
}

// handleTextCommand обрабатывает текстовую команду и вызывает соответствующие сервисы
func (s *GRPCServer) handleTextCommand(ctx context.Context, text string, sessionID string) *pb.VoiceResponse {
	// Распознаем интент из текста
	intent := s.nluEngine.RecognizeIntent(text)

	// Если распознанный интент имеет слишком низкую уверенность
	if !intent.IsConfident(model.ConfidenceThresholdLow) {
		return &pb.VoiceResponse{
			SessionId:    sessionID,
			ResponseText: "Извините, я не понимаю эту команду.",
			Successful:   false,
			Error:        "intent_not_recognized",
			RecognizedIntent: &pb.Intent{
				Name:       intent.Name,
				Confidence: intent.Confidence,
				Entities:   make(map[string]*pb.Entity),
				Parameters: make(map[string]string),
			},
		}
	}

	// Конвертируем модель интента в protobuf объект
	pbIntent := &pb.Intent{
		Name:       intent.Name,
		Confidence: intent.Confidence,
		Entities:   make(map[string]*pb.Entity),
		Parameters: intent.Parameters,
	}

	// Заполняем сущности
	for name, entity := range intent.Entities {
		pbIntent.Entities[name] = &pb.Entity{
			Value:        entity.Value,
			Confidence:   entity.Confidence,
			OriginalText: entity.OriginalText,
		}
	}

	// Выполняем действие в зависимости от типа интента
	var responseText string
	var successful bool
	var errorMsg string

	switch intent.Name {
	case model.IntentTurnOn, model.IntentTurnOff:
		responseText, successful, errorMsg = s.handleDeviceControlIntent(ctx, intent, sessionID)
	case model.IntentGetTemperature:
		responseText, successful, errorMsg = s.handleGetTemperatureIntent(ctx, intent, sessionID)
	case model.IntentSetTemperature:
		responseText, successful, errorMsg = s.handleSetTemperatureIntent(ctx, intent, sessionID)
	default:
		responseText = "Интент распознан, но обработка пока не реализована."
		successful = false
		errorMsg = "intent_handling_not_implemented"
	}

	return &pb.VoiceResponse{
		SessionId:        sessionID,
		RecognizedIntent: pbIntent,
		ResponseText:     responseText,
		Successful:       successful,
		Error:            errorMsg,
	}
}

// handleDeviceControlIntent обрабатывает интенты включения/выключения устройств
func (s *GRPCServer) handleDeviceControlIntent(ctx context.Context, intent *model.Intent, sessionID string) (string, bool, string) {
	// Получаем тип устройства из интента
	deviceEntity, deviceExists := intent.Entities[model.EntityDevice]
	if !deviceExists {
		return "Не удалось определить, какое устройство нужно " + actionVerbFromIntent(intent.Name) + ".", false, "missing_device_entity"
	}

	// Получаем комнату из интента (если есть)
	roomEntity, roomExists := intent.Entities[model.EntityRoom]
	roomDescription := ""
	if roomExists {
		roomDescription = " в " + roomEntity.Value
	}

	// В первой итерации используем упрощенную логику:
	// 1. Получаем список устройств
	// 2. Ищем первое устройство, соответствующее типу и комнате
	// 3. Отправляем команду на это устройство

	// Получаем список устройств
	devicesResp, err := s.deviceClient.ListDevices(ctx, &pb.ListDevicesRequest{
		Type:       deviceEntity.Value, // Фильтруем по типу
		OnlineOnly: true,               // Только онлайн устройства
	})
	if err != nil {
		log.Printf("Ошибка при получении списка устройств: %v", err)
		return "Не удалось получить список устройств.", false, "device_service_error"
	}

	if len(devicesResp.Devices) == 0 {
		return fmt.Sprintf("Устройство типа '%s'%s не найдено или недоступно.", deviceEntity.Value, roomDescription), false, "device_not_found"
	}

	// Ищем устройство, соответствующее комнате (если указана)
	var targetDevice *pb.Device
	if roomExists {
		for _, device := range devicesResp.Devices {
			if strings.EqualFold(device.Room, roomEntity.Value) {
				targetDevice = device
				break
			}
		}
	} else {
		// Если комната не указана, берем первое устройство из списка
		targetDevice = devicesResp.Devices[0]
	}

	if targetDevice == nil {
		return fmt.Sprintf("Устройство типа '%s'%s не найдено или недоступно.", deviceEntity.Value, roomDescription), false, "device_not_found"
	}

	// Определяем действие
	action := "turn_on"
	if intent.Name == model.IntentTurnOff {
		action = "turn_off"
	}

	// Отправляем команду на устройство
	controlResp, err := s.deviceClient.ControlDevice(ctx, &pb.ControlDeviceRequest{
		Id: targetDevice.Id,
		Command: &pb.Command{
			DeviceId:   targetDevice.Id,
			Action:     action,
			Parameters: make(map[string]string),
		},
	})
	if err != nil {
		log.Printf("Ошибка при отправке команды: %v", err)
		return fmt.Sprintf("Не удалось %s устройство.", actionVerbFromIntent(intent.Name)), false, "control_device_error"
	}

	if !controlResp.Success {
		return fmt.Sprintf("Не удалось %s устройство: %s", actionVerbFromIntent(intent.Name), controlResp.Error), false, controlResp.Error
	}

	return fmt.Sprintf("Устройство '%s'%s успешно %s.", targetDevice.Name, roomDescription, actionParticiplePastFromIntent(intent.Name)), true, ""
}

// handleGetTemperatureIntent обрабатывает интент запроса температуры
func (s *GRPCServer) handleGetTemperatureIntent(ctx context.Context, intent *model.Intent, sessionID string) (string, bool, string) {
	// Получаем комнату из интента
	roomEntity, roomExists := intent.Entities[model.EntityRoom]
	roomDescription := "дома"
	if roomExists {
		roomDescription = "в " + roomEntity.Value
	}

	// В первой итерации просто возвращаем заглушку
	// В реальной реализации нужно запросить температуру у устройств типа "sensor"
	return fmt.Sprintf("Текущая температура %s составляет 22 градуса.", roomDescription), true, ""
}

// handleSetTemperatureIntent обрабатывает интент установки температуры
func (s *GRPCServer) handleSetTemperatureIntent(ctx context.Context, intent *model.Intent, sessionID string) (string, bool, string) {
	// Получаем значение температуры из интента
	valueEntity, valueExists := intent.Entities[model.EntityValue]
	if !valueExists {
		return "Не удалось определить, какую температуру нужно установить.", false, "missing_temperature_value"
	}

	// Получаем комнату из интента
	roomEntity, roomExists := intent.Entities[model.EntityRoom]
	roomDescription := "дома"
	if roomExists {
		roomDescription = "в " + roomEntity.Value
	}

	// В первой итерации просто возвращаем заглушку
	// В реальной реализации нужно найти устройство типа "thermostat" и отправить команду
	return fmt.Sprintf("Температура %s установлена на %s градусов.", roomDescription, valueEntity.Value), true, ""
}

// actionVerbFromIntent возвращает глагол для описания действия интента
func actionVerbFromIntent(intentName string) string {
	switch intentName {
	case model.IntentTurnOn:
		return "включить"
	case model.IntentTurnOff:
		return "выключить"
	default:
		return "управлять"
	}
}

// actionParticiplePastFromIntent возвращает причастие прошедшего времени для описания результата интента
func actionParticiplePastFromIntent(intentName string) string {
	switch intentName {
	case model.IntentTurnOn:
		return "включено"
	case model.IntentTurnOff:
		return "выключено"
	default:
		return "обработано"
	}
}
