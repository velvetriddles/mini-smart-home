package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/velvetriddles/mini-smart-home/libs/kafka"
	"github.com/velvetriddles/mini-smart-home/services/api-gateway/internal/server"
)

var (
	httpPort = flag.Int("port", 8080, "HTTP server port")

	// Адреса внутренних микросервисов
	authServiceAddr   = flag.String("auth-service", "localhost:50051", "Auth service address")
	deviceServiceAddr = flag.String("device-service", "localhost:50052", "Device service address")
	voiceServiceAddr  = flag.String("voice-service", "localhost:50053", "Voice service address")

	// Настройки Kafka
	kafkaBrokers = flag.String("kafka-brokers", "localhost:9092", "Kafka brokers addresses")
	kafkaTopic   = flag.String("kafka-topic", "statusChanged", "Kafka topic for device status updates")
)

func main() {
	flag.Parse()

	log.Println("Starting API Gateway service")

	// Инициализация Kafka producer
	kafkaProducer, err := initKafkaProducer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Инициализация Kafka consumer
	kafkaConsumer, err := initKafkaConsumer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	// Подписка на сообщения из Kafka
	err = kafkaConsumer.Subscribe(func(key string, value []byte) error {
		log.Printf("Received Kafka message: key=%s, value=%s", key, string(value))
		// Здесь будет логика обработки сообщений и отправки через WebSocket
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to Kafka topic: %v", err)
	}

	// Инициализация gRPC клиента
	grpcClient, err := server.NewGRPCClient(server.GRPCConfig{
		AuthServiceAddr:   *authServiceAddr,
		DeviceServiceAddr: *deviceServiceAddr,
		VoiceServiceAddr:  *voiceServiceAddr,
	})
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer grpcClient.Close()

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

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down API Gateway service")
}

// Инициализация Kafka Producer
func initKafkaProducer() (*kafka.Producer, error) {
	brokers := []string{*kafkaBrokers}

	return kafka.NewProducer(kafka.Config{
		Brokers:  brokers,
		ClientID: "api-gateway",
		Topic:    *kafkaTopic,
	})
}

// Инициализация Kafka Consumer
func initKafkaConsumer() (*kafka.Consumer, error) {
	brokers := []string{*kafkaBrokers}

	return kafka.NewConsumer(kafka.Config{
		Brokers:  brokers,
		ClientID: "api-gateway",
		Topic:    *kafkaTopic,
	})
}
