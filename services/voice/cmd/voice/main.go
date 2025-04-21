package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/velvetriddles/mini-smart-home/proto_generated/smarthome/v1"
	"github.com/velvetriddles/mini-smart-home/services/voice/internal/server"
)

var (
	port       = flag.Int("port", 9300, "gRPC port")
	httpPort   = flag.Int("http-port", 9201, "HTTP port for metrics and health check")
	deviceAddr = flag.String("device-addr", "localhost:9200", "Device service address")
	authAddr   = flag.String("auth-addr", "localhost:9100", "Auth service address")
	logLevel   = flag.String("log-level", getEnv("LOG_LEVEL", "info"), "Уровень логирования (debug, info, warn, error)")
)

func main() {
	flag.Parse()

	// Переопределяем параметры из переменных окружения, если они заданы
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := fmt.Sscanf(envPort, "%d", port); err != nil || p != 1 {
			log.Printf("Неверный формат переменной окружения PORT: %s", envPort)
		}
	}
	if envHTTPPort := os.Getenv("HTTP_PORT"); envHTTPPort != "" {
		if p, err := fmt.Sscanf(envHTTPPort, "%d", httpPort); err != nil || p != 1 {
			log.Printf("Неверный формат переменной окружения HTTP_PORT: %s", envHTTPPort)
		}
	}
	if envDeviceAddr := os.Getenv("DEVICE_ADDR"); envDeviceAddr != "" {
		*deviceAddr = envDeviceAddr
	}
	if envAuthAddr := os.Getenv("AUTH_ADDR"); envAuthAddr != "" {
		*authAddr = envAuthAddr
	}

	// Создаем gRPC соединения с сервисами устройств и аутентификации
	deviceConn, err := grpc.Dial(*deviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к Device Service: %v", err)
	}
	defer deviceConn.Close()

	authConn, err := grpc.Dial(*authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к Auth Service: %v", err)
	}
	defer authConn.Close()

	// Инициализируем gRPC сервер
	grpcServer := grpc.NewServer()
	voiceServer := server.NewGRPCServer(deviceConn, authConn)

	// Регистрируем сервисы
	pb.RegisterVoiceServiceServer(grpcServer, voiceServer)

	// Включаем reflection для отладки с помощью таких инструментов, как grpcurl
	reflection.Register(grpcServer)

	// Регистрируем health check
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Запускаем HTTP сервер для метрик и health check
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	go func() {
		httpAddr := fmt.Sprintf(":%d", *httpPort)
		log.Printf("HTTP сервер запущен на %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, nil); err != nil {
			log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
		}
	}()

	// Запускаем gRPC сервер
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Не удалось прослушать порт %d: %v", *port, err)
	}

	log.Printf("Voice Service запущен на порту %d", *port)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Ошибка запуска gRPC сервера: %v", err)
		}
	}()

	// Ожидаем сигнала для graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Получен сигнал завершения, закрываем сервер...")

	// Останавливаем gRPC сервер с таймаутом
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("Таймаут graceful shutdown, принудительное завершение")
		grpcServer.Stop()
	case <-stopped:
		log.Println("Сервер успешно остановлен")
	}
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
