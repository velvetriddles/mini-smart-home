package model

// Intent представляет результат распознавания голосовой команды
type Intent struct {
	Name       string            // Название интента (TurnOn, TurnOff, GetTemperature и т.д.)
	Confidence float64           // Уровень уверенности в распознавании [0.0-1.0]
	Entities   map[string]Entity // Распознанные сущности (устройства, комнаты и т.д.)
	Parameters map[string]string // Дополнительные параметры команды
}

// Entity представляет распознанную сущность в команде
type Entity struct {
	Value        string  // Значение сущности (например, "лампа" или "гостиная")
	Confidence   float64 // Уровень уверенности в распознавании сущности [0.0-1.0]
	OriginalText string  // Исходный текст в команде
}

// IntentDefinition описывает поддерживаемый интент и его характеристики
type IntentDefinition struct {
	Name        string             // Название интента
	Description string             // Описание
	Examples    []string           // Примеры фраз
	Entities    []EntityDefinition // Необходимые сущности
}

// EntityDefinition описывает необходимую сущность для интента
type EntityDefinition struct {
	Name        string // Название сущности
	Description string // Описание
}

// Константы для стандартных интентов
const (
	IntentTurnOn         = "TurnOn"
	IntentTurnOff        = "TurnOff"
	IntentGetTemperature = "GetTemperature"
	IntentSetTemperature = "SetTemperature"
	IntentUnknown        = "Unknown"
)

// Константы для стандартных сущностей
const (
	EntityDevice  = "device"
	EntityRoom    = "room"
	EntityValue   = "value"
	EntityCommand = "command"
)

// Константы для уровней уверенности
const (
	ConfidenceThresholdHigh   = 0.8 // Высокий уровень уверенности
	ConfidenceThresholdMedium = 0.6 // Средний уровень уверенности
	ConfidenceThresholdLow    = 0.4 // Низкий уровень уверенности
)

// NewIntent создает новый интент с заданными параметрами
func NewIntent(name string, confidence float64) *Intent {
	return &Intent{
		Name:       name,
		Confidence: confidence,
		Entities:   make(map[string]Entity),
		Parameters: make(map[string]string),
	}
}

// AddEntity добавляет сущность к интенту
func (i *Intent) AddEntity(name, value, originalText string, confidence float64) {
	i.Entities[name] = Entity{
		Value:        value,
		Confidence:   confidence,
		OriginalText: originalText,
	}
}

// AddParameter добавляет параметр к интенту
func (i *Intent) AddParameter(key, value string) {
	i.Parameters[key] = value
}

// IsConfident возвращает true, если уровень уверенности выше или равен порогу
func (i *Intent) IsConfident(threshold float64) bool {
	return i.Confidence >= threshold
}

// UnknownIntent создает стандартный "неизвестный" интент с низким уровнем уверенности
func UnknownIntent() *Intent {
	return NewIntent(IntentUnknown, 0.1)
}
