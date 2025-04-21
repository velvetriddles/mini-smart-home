package internal

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	smarthomev1 "github.com/velvetriddles/mini-smart-home/proto_generated/smarthome/v1"
)

// WebSocketConfig содержит настройки для WebSocket-прокси
type WebSocketConfig struct {
	DeviceClient smarthomev1.DeviceServiceClient
}

// NewWebSocketProxy создает новый обработчик для WebSocket соединений
func NewWebSocketProxy(config WebSocketConfig) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // В продакшене следует ограничить источники запроса
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Улучшаем соединение до WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading connection to WebSocket: %v", err)
			return
		}
		defer conn.Close()

		log.Printf("WebSocket connection established from %s", r.RemoteAddr)

		// Создаем контекст, который можно отменить при закрытии соединения
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Устанавливаем gRPC-стрим для получения обновлений статусов устройств
		stream, err := config.DeviceClient.StreamStatuses(ctx)
		if err != nil {
			log.Printf("Error establishing gRPC stream: %v", err)
			return
		}

		// Отправляем запрос на подписку на все устройства
		err = stream.Send(&smarthomev1.StatusRequest{
			SubscribeAll: true,
		})
		if err != nil {
			log.Printf("Error sending subscription request: %v", err)
			return
		}

		// Запускаем чтение из WebSocket в отдельной горутине
		// В данной реализации мы просто считываем сообщения, но игнорируем их
		go func() {
			defer cancel() // отменяем контекст при завершении чтения
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err,
						websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket read error: %v", err)
					}
					break
				}
			}
		}()

		// Читаем из gRPC-стрима и отправляем данные в WebSocket
		for {
			statusUpdate, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Printf("Error receiving from gRPC stream: %v", err)
				}
				break
			}

			// Отправляем обновление состояния устройства клиенту через WebSocket
			if err := conn.WriteJSON(statusUpdate); err != nil {
				log.Printf("Error sending message to WebSocket: %v", err)
				break
			}
		}

		log.Printf("WebSocket connection closed for %s", r.RemoteAddr)
	}
}
