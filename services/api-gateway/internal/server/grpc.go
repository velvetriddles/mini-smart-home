package server

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCConfig содержит конфигурацию для gRPC-клиентов
type GRPCConfig struct {
	AuthServiceAddr   string
	DeviceServiceAddr string
	VoiceServiceAddr  string
}

// GRPCClient представляет gRPC соединения к внутренним микросервисам
type GRPCClient struct {
	authConn   *grpc.ClientConn
	deviceConn *grpc.ClientConn
	voiceConn  *grpc.ClientConn
}

// NewGRPCClient создает новый gRPC клиент для взаимодействия с микросервисами
func NewGRPCClient(cfg GRPCConfig) (*GRPCClient, error) {
	// Опции подключения к gRPC-серверам
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Подключение к сервису auth
	authConn, err := grpc.Dial(cfg.AuthServiceAddr, opts...)
	if err != nil {
		return nil, err
	}

	// Подключение к сервису device
	deviceConn, err := grpc.Dial(cfg.DeviceServiceAddr, opts...)
	if err != nil {
		return nil, err
	}

	// Подключение к сервису voice
	voiceConn, err := grpc.Dial(cfg.VoiceServiceAddr, opts...)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		authConn:   authConn,
		deviceConn: deviceConn,
		voiceConn:  voiceConn,
	}, nil
}

// Close закрывает все gRPC-соединения
func (c *GRPCClient) Close() {
	if c.authConn != nil {
		if err := c.authConn.Close(); err != nil {
			log.Printf("Failed to close auth connection: %v", err)
		}
	}
	
	if c.deviceConn != nil {
		if err := c.deviceConn.Close(); err != nil {
			log.Printf("Failed to close device connection: %v", err)
		}
	}
	
	if c.voiceConn != nil {
		if err := c.voiceConn.Close(); err != nil {
			log.Printf("Failed to close voice connection: %v", err)
		}
	}
} 