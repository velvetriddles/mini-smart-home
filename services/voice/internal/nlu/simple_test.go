package nlu

import (
	"testing"

	"github.com/velvetriddles/mini-smart-home/services/voice/internal/model"
)

func TestSimpleNLU_RecognizeIntent(t *testing.T) {
	nlu := NewSimpleNLU()

	tests := []struct {
		name           string
		text           string
		expectedIntent string
		minConfidence  float64
		entities       map[string]string
	}{
		{
			name:           "TurnOn_LampInLivingRoom",
			text:           "Включи лампу в гостиной",
			expectedIntent: model.IntentTurnOn,
			minConfidence:  0.6,
			entities: map[string]string{
				model.EntityDevice: "лампа",
				model.EntityRoom:   "гостиная",
			},
		},
		{
			name:           "TurnOn_Light",
			text:           "Зажги свет на кухне",
			expectedIntent: model.IntentTurnOn,
			minConfidence:  0.6,
			entities: map[string]string{
				model.EntityDevice: "свет",
				model.EntityRoom:   "кухня",
			},
		},
		{
			name:           "TurnOff_LampInLivingRoom",
			text:           "Выключи лампу в гостиной",
			expectedIntent: model.IntentTurnOff,
			minConfidence:  0.6,
			entities: map[string]string{
				model.EntityDevice: "лампа",
				model.EntityRoom:   "гостиная",
			},
		},
		{
			name:           "TurnOff_Light",
			text:           "Погаси свет в спальне",
			expectedIntent: model.IntentTurnOff,
			minConfidence:  0.6,
			entities: map[string]string{
				model.EntityDevice: "свет",
				model.EntityRoom:   "спальня",
			},
		},
		{
			name:           "GetTemperature",
			text:           "Какая температура в спальне",
			expectedIntent: model.IntentGetTemperature,
			minConfidence:  0.6,
			entities: map[string]string{
				model.EntityRoom: "спальня",
			},
		},
		{
			name:           "GetTemperature_Alternative",
			text:           "Сколько градусов на кухне",
			expectedIntent: model.IntentGetTemperature,
			minConfidence:  0.6,
			entities: map[string]string{
				model.EntityRoom: "кухня",
			},
		},
		{
			name:           "SetTemperature",
			text:           "Установи температуру 22 градуса в спальне",
			expectedIntent: model.IntentSetTemperature,
			minConfidence:  0.6,
			entities: map[string]string{
				model.EntityValue: "22",
				model.EntityRoom:  "спальня",
			},
		},
		{
			name:           "SetTemperature_Alternative",
			text:           "Поставь температуру 25 градусов на кухне",
			expectedIntent: model.IntentSetTemperature,
			minConfidence:  0.6,
			entities: map[string]string{
				model.EntityValue: "25",
				model.EntityRoom:  "кухня",
			},
		},
		{
			name:           "Unknown_Intent",
			text:           "Что такое умный дом",
			expectedIntent: model.IntentUnknown,
			minConfidence:  0.0, // Неизвестный интент имеет низкую уверенность
			entities:       map[string]string{},
		},
		{
			name:           "TurnOn_MissingDevice",
			text:           "Включи что-нибудь в гостиной",
			expectedIntent: model.IntentTurnOn,
			minConfidence:  0.0, // Уверенность низкая, так как устройство не распознано
			entities: map[string]string{
				model.EntityRoom: "гостиная",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent := nlu.RecognizeIntent(tt.text)

			if intent.Name != tt.expectedIntent {
				t.Errorf("Expected intent %s, got %s", tt.expectedIntent, intent.Name)
			}

			if intent.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %f, got %f", tt.minConfidence, intent.Confidence)
			}

			for entityKey, expectedValue := range tt.entities {
				entity, exists := intent.Entities[entityKey]
				if !exists {
					t.Errorf("Expected entity %s not found", entityKey)
					continue
				}

				if entity.Value != expectedValue {
					t.Errorf("Expected entity %s value to be %s, got %s", entityKey, expectedValue, entity.Value)
				}
			}
		})
	}
}

func TestSimpleNLU_GetSupportedIntents(t *testing.T) {
	nlu := NewSimpleNLU()
	intents := nlu.GetSupportedIntents()

	expectedCount := 4 // Должно быть 4 поддерживаемых интента
	if len(intents) != expectedCount {
		t.Errorf("Expected %d supported intents, got %d", expectedCount, len(intents))
	}

	// Проверка наличия всех необходимых интентов
	intentNames := map[string]bool{}
	for _, intent := range intents {
		intentNames[intent.Name] = true
	}

	requiredIntents := []string{
		model.IntentTurnOn,
		model.IntentTurnOff,
		model.IntentGetTemperature,
		model.IntentSetTemperature,
	}

	for _, name := range requiredIntents {
		if !intentNames[name] {
			t.Errorf("Required intent %s not found in supported intents", name)
		}
	}
}
