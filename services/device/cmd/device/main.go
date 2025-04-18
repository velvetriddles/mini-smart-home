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
	"github.com/velvetriddles/mini-smart-home/services/device/internal/datastore"
	"github.com/velvetriddles/mini-smart-home/services/device/internal/server"
	pb "github.com/velvetriddles/mini-smart-home/services/device/proto/smarthome/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	grpcPort = flag.Int("port", 9200, "GRPC server port")
	httpPort = flag.Int("http-port", 9101, "HTTP metrics port")
)

func main() {
	flag.Parse()

	// Инициализация in-memory хранилища устройств
	store := datastore.NewMemoryStore()

	// Добавляем тестовые устройства для разработки
	store.AddTestDevices()

	// Настраиваем gRPC сервер
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", *grpcPort, err)
	}

	// Создаем gRPC сервер
	grpcServer := grpc.NewServer()

	// Регистрируем Device Service
	deviceService := server.NewGRPCServer(store)
	pb.RegisterDeviceServiceServer(grpcServer, deviceService)

	// Регистрируем Health Service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Включаем reflection для отладки с помощью grpcurl
	reflection.Register(grpcServer)

	// Запускаем HTTP сервер с метриками Prometheus
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Запускаем HTTP сервер в отдельной горутине
	go func() {
		log.Printf("Starting HTTP metrics server on :%d", *httpPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Запускаем gRPC сервер в отдельной горутине
	go func() {
		log.Printf("Starting gRPC server on :%d", *grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Ожидаем сигналы для корректного завершения
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	// Корректно останавливаем серверы
	log.Println("Stopping servers...")

	// Даем серверу 5 секунд на корректное завершение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	grpcServer.GracefulStop()

	<-ctx.Done()
	log.Println("Device service shutdown complete")
}
