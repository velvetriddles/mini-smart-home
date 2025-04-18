.PHONY: build run-device test-device docker-build docker-run clean

# Параметры
APP_NAME = device
BUILD_DIR = ./bin
DOCKER_IMAGE = smarthome/device:latest
GRPC_PORT = 9200
HTTP_PORT = 9101

# Сборка бинарного файла
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/device

# Запуск сервиса локально
run-device: build
	@echo "Running $(APP_NAME) service..."
	$(BUILD_DIR)/$(APP_NAME) --port=$(GRPC_PORT) --http-port=$(HTTP_PORT)

# Запуск тестов
test-device:
	@echo "Running tests for $(APP_NAME)..."
	go test -v ./...

# Запуск с покрытием кода тестами
test-coverage:
	@echo "Running tests with coverage for $(APP_NAME)..."
	go test -v -cover ./...

# Статический анализ кода
lint:
	@echo "Running linter for $(APP_NAME)..."
	golangci-lint run ./...

# Сборка Docker образа
docker-build: build
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .

# Запуск Docker контейнера
docker-run:
	@echo "Running Docker container $(DOCKER_IMAGE)..."
	docker run -p $(GRPC_PORT):$(GRPC_PORT) -p $(HTTP_PORT):$(HTTP_PORT) $(DOCKER_IMAGE)

# Очистка артефактов сборки
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Cleaned!"

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  run-device   - Run the device service locally"
	@echo "  test-device  - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint         - Run golangci-lint"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help" 