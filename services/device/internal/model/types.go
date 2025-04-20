package model

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CommandResult представляет результат выполнения команды
type CommandResult struct {
	Success   bool                   // Успешность выполнения
	Message   string                 // Текстовое сообщение
	Timestamp *timestamppb.Timestamp // Метка времени результата
}
