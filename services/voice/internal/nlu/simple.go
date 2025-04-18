package nlu

import (
	"regexp"
	"strings"

	"github.com/velvetriddles/mini-smart-home/services/voice/internal/model"
)

// SimpleNLU представляет простой механизм распознавания голосовых команд на основе ключевых слов
type SimpleNLU struct {
	deviceTypes []string
	rooms       []string
}

// NewSimpleNLU создает новый экземпляр SimpleNLU с предварительно настроенными распознаваемыми значениями
func NewSimpleNLU() *SimpleNLU {
	return &SimpleNLU{
		deviceTypes: []string{"лампа", "лампу", "свет", "света", "розетка", "розетку", "чайник", "чайника",
			"кондиционер", "кондиционера", "вентилятор", "вентилятора", "телевизор", "телевизора"},
		rooms: []string{"кухня", "кухне", "гостиная", "гостиной", "спальня", "спальне", "ванная", "ванной",
			"коридор", "коридоре", "туалет", "туалете", "прихожая", "прихожей"},
	}
}

// RecognizeIntent анализирует текстовый ввод и определяет интент
func (s *SimpleNLU) RecognizeIntent(text string) *model.Intent {
	// Приводим к нижнему регистру для упрощения обработки
	normalized := strings.ToLower(text)

	// Проверяем команду "включить"
	if containsAny(normalized, []string{"включи", "включить", "зажги", "зажечь", "запусти", "активируй", "вруби"}) {
		return s.handleTurnOnIntent(normalized)
	}

	// Проверяем команду "выключить"
	if containsAny(normalized, []string{"выключи", "выключить", "погаси", "потуши", "отключи", "деактивируй", "вырубай"}) {
		return s.handleTurnOffIntent(normalized)
	}

	// Проверяем запрос температуры
	if containsAny(normalized, []string{"какая температура", "сколько градусов", "какая сейчас температура"}) {
		return s.handleGetTemperatureIntent(normalized)
	}

	// Проверяем команду установки температуры
	if containsAny(normalized, []string{"установи температуру", "поставь температуру", "сделай температуру"}) {
		return s.handleSetTemperatureIntent(normalized)
	}

	// Если ничего не распознано
	return model.UnknownIntent()
}

// handleTurnOnIntent обрабатывает интент включения устройства
func (s *SimpleNLU) handleTurnOnIntent(text string) *model.Intent {
	// Создаем интент с начальной уверенностью
	intent := model.NewIntent(model.IntentTurnOn, 0.7)

	// Ищем тип устройства
	deviceType, deviceConfidence := s.extractDeviceType(text)
	if deviceType != "" {
		intent.AddEntity(model.EntityDevice, deviceType, deviceType, deviceConfidence)
		intent.Confidence = (intent.Confidence + deviceConfidence) / 2
	} else {
		// Если устройство не найдено, снижаем уверенность
		intent.Confidence = 0.3
	}

	// Ищем комнату
	room, roomConfidence := s.extractRoom(text)
	if room != "" {
		intent.AddEntity(model.EntityRoom, room, room, roomConfidence)
	}

	// Добавляем параметр действия
	intent.AddParameter("action", "turn_on")

	return intent
}

// handleTurnOffIntent обрабатывает интент выключения устройства
func (s *SimpleNLU) handleTurnOffIntent(text string) *model.Intent {
	// Создаем интент с начальной уверенностью
	intent := model.NewIntent(model.IntentTurnOff, 0.7)

	// Ищем тип устройства
	deviceType, deviceConfidence := s.extractDeviceType(text)
	if deviceType != "" {
		intent.AddEntity(model.EntityDevice, deviceType, deviceType, deviceConfidence)
		intent.Confidence = (intent.Confidence + deviceConfidence) / 2
	} else {
		// Если устройство не найдено, снижаем уверенность
		intent.Confidence = 0.3
	}

	// Ищем комнату
	room, roomConfidence := s.extractRoom(text)
	if room != "" {
		intent.AddEntity(model.EntityRoom, room, room, roomConfidence)
	}

	// Добавляем параметр действия
	intent.AddParameter("action", "turn_off")

	return intent
}

// handleGetTemperatureIntent обрабатывает интент запроса температуры
func (s *SimpleNLU) handleGetTemperatureIntent(text string) *model.Intent {
	// Создаем интент с начальной уверенностью
	intent := model.NewIntent(model.IntentGetTemperature, 0.7)

	// Ищем комнату
	room, roomConfidence := s.extractRoom(text)
	if room != "" {
		intent.AddEntity(model.EntityRoom, room, room, roomConfidence)
		intent.Confidence = (intent.Confidence + roomConfidence) / 2
	} else {
		// Если комната не найдена, используем значение по умолчанию
		intent.AddEntity(model.EntityRoom, "общая", "общая", 0.5)
	}

	return intent
}

// handleSetTemperatureIntent обрабатывает интент установки температуры
func (s *SimpleNLU) handleSetTemperatureIntent(text string) *model.Intent {
	// Создаем интент с начальной уверенностью
	intent := model.NewIntent(model.IntentSetTemperature, 0.7)

	// Ищем значение температуры
	value, valueConfidence := s.extractTemperatureValue(text)
	if value != "" {
		intent.AddEntity(model.EntityValue, value, value, valueConfidence)
		intent.Confidence = (intent.Confidence + valueConfidence) / 2
	} else {
		// Если значение не найдено, снижаем уверенность
		intent.Confidence = 0.3
	}

	// Ищем комнату
	room, roomConfidence := s.extractRoom(text)
	if room != "" {
		intent.AddEntity(model.EntityRoom, room, room, roomConfidence)
	}

	return intent
}

// extractDeviceType извлекает тип устройства из текста
func (s *SimpleNLU) extractDeviceType(text string) (string, float64) {
	for _, device := range s.deviceTypes {
		if strings.Contains(text, device) {
			// Нормализуем в именительный падеж для единообразия
			switch device {
			case "лампу", "розетку":
				return device[:len(device)-1] + "а", 0.9
			case "света":
				return "свет", 0.9
			case "чайника", "кондиционера", "вентилятора", "телевизора":
				return device[:len(device)-1], 0.9
			default:
				return device, 0.9
			}
		}
	}
	return "", 0.0
}

// extractRoom извлекает название комнаты из текста
func (s *SimpleNLU) extractRoom(text string) (string, float64) {
	for _, room := range s.rooms {
		if strings.Contains(text, room) {
			// Нормализуем в именительный падеж для единообразия
			switch room {
			case "кухне", "гостиной", "спальне", "ванной", "коридоре", "туалете", "прихожей":
				if room == "кухне" {
					return "кухня", 0.9
				} else if room == "ванной" {
					return "ванная", 0.9
				} else if room == "гостиной" {
					return "гостиная", 0.9
				} else if room == "спальне" {
					return "спальня", 0.9
				} else if room == "коридоре" {
					return "коридор", 0.9
				} else if room == "туалете" {
					return "туалет", 0.9
				} else if room == "прихожей" {
					return "прихожая", 0.9
				}
			}
			return room, 0.9
		}
	}
	return "", 0.0
}

// extractTemperatureValue извлекает значение температуры из текста
func (s *SimpleNLU) extractTemperatureValue(text string) (string, float64) {
	// Используем регулярное выражение для поиска числовых значений
	re := regexp.MustCompile(`(\d+)( градусов| градуса| градус)?`)
	matches := re.FindStringSubmatch(text)

	if len(matches) > 1 {
		return matches[1], 0.9
	}

	return "", 0.0
}

// containsAny проверяет наличие любой из строк списка в исходной строке
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// GetSupportedIntents возвращает список поддерживаемых интентов с описаниями
func (s *SimpleNLU) GetSupportedIntents() []model.IntentDefinition {
	return []model.IntentDefinition{
		{
			Name:        model.IntentTurnOn,
			Description: "Включение устройства",
			Examples:    []string{"Включи свет в гостиной", "Зажги лампу на кухне", "Включи чайник"},
			Entities: []model.EntityDefinition{
				{Name: model.EntityDevice, Description: "Устройство для включения"},
				{Name: model.EntityRoom, Description: "Комната, где находится устройство"},
			},
		},
		{
			Name:        model.IntentTurnOff,
			Description: "Выключение устройства",
			Examples:    []string{"Выключи свет в гостиной", "Погаси лампу на кухне", "Отключи чайник"},
			Entities: []model.EntityDefinition{
				{Name: model.EntityDevice, Description: "Устройство для выключения"},
				{Name: model.EntityRoom, Description: "Комната, где находится устройство"},
			},
		},
		{
			Name:        model.IntentGetTemperature,
			Description: "Запрос текущей температуры",
			Examples:    []string{"Какая температура в спальне", "Сколько градусов на кухне"},
			Entities: []model.EntityDefinition{
				{Name: model.EntityRoom, Description: "Комната, где нужно узнать температуру"},
			},
		},
		{
			Name:        model.IntentSetTemperature,
			Description: "Установка температуры",
			Examples:    []string{"Установи температуру 22 градуса в спальне", "Сделай 25 градусов на кухне"},
			Entities: []model.EntityDefinition{
				{Name: model.EntityValue, Description: "Значение температуры"},
				{Name: model.EntityRoom, Description: "Комната, где нужно установить температуру"},
			},
		},
	}
}
