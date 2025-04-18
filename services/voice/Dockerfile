FROM golang:1.22-bookworm AS builder

WORKDIR /app

# Копируем только модули для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /voice-service ./cmd/voice/main.go

# Конечный образ
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Копируем бинарный файл из предыдущего этапа
COPY --from=builder /voice-service /app/voice-service

EXPOSE 9300
EXPOSE 9201

# Запускаем приложение
ENTRYPOINT ["/app/voice-service"]