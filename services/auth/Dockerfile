FROM golang:1.20-alpine AS builder

WORKDIR /app

# Копирование и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o auth .

# Финальный образ
FROM alpine:3.17

WORKDIR /app

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata

# Копирование бинарных файлов и SQL скриптов из предыдущего этапа
COPY --from=builder /app/auth .
COPY --from=builder /app/sql ./sql/

# Установка переменных среды по умолчанию
ENV POSTGRES_DSN="postgres://postgres:postgres@postgres:5432/smarthome?sslmode=disable" \
    REDIS_DSN="redis://redis:6379/0" \
    JWT_SECRET="change-this-in-production" \
    JWT_TTL="24h" \
    GRPC_PORT="50051" \
    HTTP_PORT="9090" \
    LOG_LEVEL="info"

# Проверка жизнеспособности
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:${HTTP_PORT}/healthz || exit 1

# Открываем порты для gRPC и HTTP серверов
EXPOSE ${GRPC_PORT} ${HTTP_PORT}

# Запуск приложения
CMD ["./auth"] 