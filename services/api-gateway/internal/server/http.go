package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	smarthomev1 "github.com/velvetriddles/mini-smart-home/proto_generated/smarthome/v1"
	"github.com/velvetriddles/mini-smart-home/services/api-gateway/internal"
	authMiddleware "github.com/velvetriddles/mini-smart-home/services/api-gateway/internal/middleware"
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
	router     *chi.Mux
	config     HTTPConfig
	authConn   *grpc.ClientConn
	deviceConn *grpc.ClientConn
	voiceConn  *grpc.ClientConn
	upgrader   websocket.Upgrader
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

	// Добавляем JWT middleware, но не для всех маршрутов
	// Переместили в setupRoutes

	// Соединение с gRPC-сервисами
	server.setupGRPCConnections()

	// Настройка маршрутов
	server.setupRoutes()

	return server
}

// Устанавливает соединения с gRPC-сервисами
func (s *HTTPServer) setupGRPCConnections() {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	var err error

	// Соединение с Auth Service
	s.authConn, err = grpc.Dial(s.config.AuthServiceAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to Auth Service: %v", err)
	}

	// Соединение с Device Service
	s.deviceConn, err = grpc.Dial(s.config.DeviceServiceAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to Device Service: %v", err)
	}

	// Соединение с Voice Service
	s.voiceConn, err = grpc.Dial(s.config.VoiceServiceAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to Voice Service: %v", err)
	}
}

// Настраивает маршруты HTTP
func (s *HTTPServer) setupRoutes() {
	// Публичные маршруты, не требующие авторизации
	s.router.Group(func(r chi.Router) {
		// Обработчик проверки здоровья
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("OK"))
		})

		// WebSocket для получения статусов устройств
		deviceClient := smarthomev1.NewDeviceServiceClient(s.deviceConn)
		r.Get("/ws/status", internal.NewWebSocketProxy(internal.WebSocketConfig{
			DeviceClient: deviceClient,
		}))

		// Публичные API, требующие авторизации
		r.Group(func(r chi.Router) {
			// Подключаем gRPC-gateway для публичных эндпоинтов
			// Например, авторизация
			ctx := context.Background()
			mux := runtime.NewServeMux(
				runtime.WithIncomingHeaderMatcher(func(k string) (string, bool) {
					if strings.ToLower(k) == "authorization" {
						return k, true
					}
					return runtime.DefaultHeaderMatcher(k)
				}),
			)

			opts := []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			}

			// Регистрируем только AuthService для публичных маршрутов
			if err := smarthomev1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, s.config.AuthServiceAddr, opts); err != nil {
				log.Fatalf("Failed to register public AuthService handler: %v", err)
			}

			r.Mount("/api/v1/auth", http.StripPrefix("/api/v1/auth", mux))
		})
	})

	// Защищенные маршруты, требующие авторизации
	s.router.Group(func(r chi.Router) {
		// Добавляем JWT middleware для защищенных маршрутов
		r.Use(authMiddleware.JWT)

		// Настраиваем защищенные gRPC-gateway эндпоинты
		ctx := context.Background()
		mux := runtime.NewServeMux()

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		// Регистрируем Device и Voice сервисы для защищенных маршрутов
		if err := smarthomev1.RegisterDeviceServiceHandlerFromEndpoint(ctx, mux, s.config.DeviceServiceAddr, opts); err != nil {
			log.Fatalf("Failed to register DeviceService handler: %v", err)
		}

		if err := smarthomev1.RegisterVoiceServiceHandlerFromEndpoint(ctx, mux, s.config.VoiceServiceAddr, opts); err != nil {
			log.Fatalf("Failed to register VoiceService handler: %v", err)
		}

		r.Mount("/api/v1", http.StripPrefix("/api/v1", mux))
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

// Start запускает HTTP-сервер
func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Printf("Starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// Close закрывает все соединения
func (s *HTTPServer) Close() {
	if s.authConn != nil {
		s.authConn.Close()
	}
	if s.deviceConn != nil {
		s.deviceConn.Close()
	}
	if s.voiceConn != nil {
		s.voiceConn.Close()
	}
}
