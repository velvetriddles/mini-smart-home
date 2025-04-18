package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	authMiddleware "github.com/velvetriddles/mini-smart-home/services/api-gateway/internal/middleware"
	smarthomev1 "github.com/velvetriddles/mini-smart-home/services/api-gateway/proto/smarthome/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config содержит конфигурацию для HTTP-сервера
type HTTPConfig struct {
	Port              int
	AuthServiceAddr   string
	DeviceServiceAddr string
	VoiceServiceAddr  string
}

// Server представляет собой HTTP-сервер
type HTTPServer struct {
	router   *chi.Mux
	config   HTTPConfig
	grpcConn *grpc.ClientConn
	upgrader websocket.Upgrader
}

// Создает новый HTTP-сервер
func NewHTTPServer(config HTTPConfig) *HTTPServer {
	r := chi.NewRouter()

	// Настройка websocket upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // В продакшне следует ограничить источники запроса
		},
	}

	server := &HTTPServer{
		router:   r,
		config:   config,
		upgrader: upgrader,
	}

	// Настройка middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Добавляем JWT middleware
	r.Use(jwtauth.Verifier(authMiddleware.TokenAuth))
	r.Use(jwtauth.Authenticator)

	// Настройка маршрутов
	server.setupRoutes()

	return server
}

// Настраивает маршруты HTTP
func (s *HTTPServer) setupRoutes() {
	// Подключаем gRPC-gateway
	s.setupGRPCGateway()

	// Обработчик WebSocket
	s.router.Get("/ws", s.handleWebSocket)

	// Обработчик проверки здоровья
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Тестовый обработчик для генерации токена (для отладки)
	s.router.Post("/debug/token", func(w http.ResponseWriter, r *http.Request) {
		token, err := authMiddleware.GenerateToken("test-user-id", "testuser")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	})
}

// Настраивает gRPC-gateway для проксирования REST запросов в gRPC
func (s *HTTPServer) setupGRPCGateway() {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := smarthomev1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, s.config.AuthServiceAddr, opts); err != nil {
		log.Fatalf("Failed to register AuthService handler: %v", err)
	}

	if err := smarthomev1.RegisterDeviceServiceHandlerFromEndpoint(ctx, mux, s.config.DeviceServiceAddr, opts); err != nil {
		log.Fatalf("Failed to register DeviceService handler: %v", err)
	}

	if err := smarthomev1.RegisterVoiceServiceHandlerFromEndpoint(ctx, mux, s.config.VoiceServiceAddr, opts); err != nil {
		log.Fatalf("Failed to register VoiceService handler: %v", err)
	}

	s.router.Mount("/api/v1", http.StripPrefix("/api/v1", mux))
}

// Обработчик WebSocket соединений
func (s *HTTPServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Улучшаем соединение до WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Простой эхо-сервер для WebSocket
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			break
		}

		// Эхо-ответ
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Error writing WebSocket message: %v", err)
			break
		}
	}
}

// Start запускает HTTP-сервер
func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Printf("Starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}
