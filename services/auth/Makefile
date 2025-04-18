# Makefile для Auth Service

# Переменные
BINARY_NAME=auth
GO=go
GOFLAGS=-ldflags="-s -w"
DOCKER_IMAGE=smarthome/auth-service
DOCKER_TAG=latest

# Порты
GRPC_PORT=50051
HTTP_PORT=9090

# Переменные окружения для тестирования
TEST_JWT_SECRET=test-jwt-secret
TEST_POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/smarthome_test?sslmode=disable
TEST_REDIS_DSN=redis://localhost:6379/1

.PHONY: all build clean run-auth test-auth test-jwt init-db docker-build docker-run lint

all: build

# Сборка сервиса
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

# Очистка
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

# Проверка наличия необходимых переменных окружения
check-env:
	@if [ -z "$(POSTGRES_DSN)" ]; then \
		echo "POSTGRES_DSN is not set. Using default."; \
		export POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/smarthome?sslmode=disable"; \
	fi
	@if [ -z "$(REDIS_DSN)" ]; then \
		echo "REDIS_DSN is not set. Using default."; \
		export REDIS_DSN="redis://localhost:6379/0"; \
	fi
	@if [ -z "$(JWT_SECRET)" ]; then \
		echo "JWT_SECRET is not set. Using default."; \
		export JWT_SECRET="super-secret-key-change-in-production"; \
	fi

# Запуск локального сервера
run-auth:
	POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/smarthome?sslmode=disable" \
	REDIS_DSN="redis://localhost:6379/0" \
	JWT_SECRET="super-secret-key-change-in-production" \
	JWT_TTL="24h" \
	GRPC_PORT="$(GRPC_PORT)" \
	HTTP_PORT="$(HTTP_PORT)" \
	$(GO) run .

# Запуск полного набора тестов
test-auth:
	POSTGRES_DSN="$(TEST_POSTGRES_DSN)" \
	REDIS_DSN="$(TEST_REDIS_DSN)" \
	JWT_SECRET="$(TEST_JWT_SECRET)" \
	$(GO) test -v -cover ./...

# Запуск теста функции GenerateJWT
test-jwt:
	JWT_SECRET="$(TEST_JWT_SECRET)" \
	$(GO) test -v -run "TestGenerate(JTI|JWT)"

# Запуск тестов валидации и отзыва токенов
test-token:
	JWT_SECRET="$(TEST_JWT_SECRET)" \
	$(GO) test -v -run "TestValidate|TestRevoke"

# Создание тестовой базы данных
init-test-db:
	PGPASSWORD=postgres psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS smarthome_test;"
	PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE smarthome_test;"
	PGPASSWORD=postgres psql -h localhost -U postgres -d smarthome_test -f sql/schema.sql

# Создание prod базы данных
init-db:
	PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE IF NOT EXISTS smarthome;"
	PGPASSWORD=postgres psql -h localhost -U postgres -d smarthome -f sql/schema.sql

# Сборка Docker образа
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Запуск Docker контейнера
docker-run:
	docker run -p $(GRPC_PORT):$(GRPC_PORT) -p $(HTTP_PORT):$(HTTP_PORT) \
		-e POSTGRES_DSN="$(POSTGRES_DSN)" \
		-e REDIS_DSN="$(REDIS_DSN)" \
		-e JWT_SECRET="$(JWT_SECRET)" \
		-e JWT_TTL="24h" \
		-e GRPC_PORT="$(GRPC_PORT)" \
		-e HTTP_PORT="$(HTTP_PORT)" \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Запуск lint проверки
lint:
	golangci-lint run ./... 