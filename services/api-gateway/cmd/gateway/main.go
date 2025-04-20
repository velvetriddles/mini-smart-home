package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/velvetriddles/mini-smart-home/services/api-gateway/internal/server"
)

var (
	httpPort = flag.Int("port", 8080, "HTTP server port")

	// Адреса внутренних микросервисов
	authServiceAddr   = flag.String("auth-service", "localhost:50051", "Auth service address")
	deviceServiceAddr = flag.String("device-service", "localhost:50052", "Device service address")
	voiceServiceAddr  = flag.String("voice-service", "localhost:50053", "Voice service address")
)

func main() {
	flag.Parse()

	log.Println("Starting API Gateway service")

	// Инициализация HTTP сервера
	httpServer := server.NewHTTPServer(server.HTTPConfig{
		Port:              *httpPort,
		AuthServiceAddr:   *authServiceAddr,
		DeviceServiceAddr: *deviceServiceAddr,
		VoiceServiceAddr:  *voiceServiceAddr,
	})

	// Запуск HTTP сервера в горутине
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	log.Printf("API Gateway running on port %d", *httpPort)
	log.Printf("Connected to services: Auth=%s, Device=%s, Voice=%s",
		*authServiceAddr, *deviceServiceAddr, *voiceServiceAddr)

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log.Printf("Received signal %v, shutting down...", sig)

	// Закрытие соединений
	httpServer.Close()

	log.Println("API Gateway service stopped")
}
