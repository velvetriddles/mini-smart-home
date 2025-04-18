.PHONY: build run-voice test-voice docker-build docker-run clean

# Переменные
BINARY_NAME=voice-service
DOCKER_TAG=smarthome/voice-service:latest

# Сборка бинарного файла
build:
	go build -o $(BINARY_NAME) ./cmd/voice/main.go

# Запуск сервиса
run-voice:
	go run ./cmd/voice/main.go

# Запуск тестов
test-voice:
	go test -v ./...

# Сборка Docker-образа
docker-build:
	docker build -t $(DOCKER_TAG) .

# Запуск Docker-контейнера
docker-run:
	docker run -p 9300:9300 -p 9201:9201 $(DOCKER_TAG)

# Очистка
clean:
	go clean
	rm -f $(BINARY_NAME) 