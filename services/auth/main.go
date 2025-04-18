package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Auth service starting...")

	// Получение конфигурации из переменных окружения
	config := &Config{
		GrpcPort:    getEnv("GRPC_PORT", "50051"),
		HttpPort:    getEnv("HTTP_PORT", "9090"),
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/smarthome?sslmode=disable"),
		RedisDSN:    getEnv("REDIS_DSN", "redis://localhost:6379/0"),
		JwtSecret:   getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
		JwtTTL:      getEnvDuration("JWT_TTL", 24*time.Hour),
	}

	// Создание и запуск сервера
	server, err := NewServer(config)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	// Запуск в горутине
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("Auth service started",
		zap.String("grpc_port", config.GrpcPort),
		zap.String("http_port", config.HttpPort))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Auth service stopped gracefully")
}

// getEnv получает переменную окружения с значением по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvDuration получает переменную окружения в формате Duration с значением по умолчанию
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
