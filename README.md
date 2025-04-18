# Smart-Home

Монорепозиторий проекта Smart-Home (Дипломная версия).

## Структура проекта

- `services/` - микросервисы проекта
  - `api-gateway/` - API-шлюз для внешних запросов (REST+WS)
  - `auth/` - сервис аутентификации
  - `device/` - сервис управления устройствами
  - `voice/` - сервис голосового управления
  - `web/` - SPA-сервис
- `libs/` - общие библиотеки
  - `kafka/` - обертка над Sarama для работы с Kafka
- `proto/` - контракты gRPC собственной разработки
  - `smarthome/` - API системы "Умный дом"
- `third_party/` - сторонние proto-файлы и зависимости
  - `google/` - локальные копии Google API proto-файлов
- `web/` - фронтенд-приложение (React+Vite)
- `infra/` - инфраструктурные конфигурации
- `docs/` - документация
- `.github/` - GitHub Actions конфигурации

## Транспорты

- gRPC внутри микросервисной архитектуры
- REST+WebSockets для внешних клиентов

## Шины

- Kafka (topics: `commands`, `statusChanged`)
- MQTT (topics: `wb/device/+/status`, `wb/device/+/command`)

## Базы данных

- PostgreSQL (device_state, users)
- Redis (revoked-jwt)

## Генерация кода из proto-файлов

Генерация кода осуществляется с помощью protoc и скрипта generate_proto.sh:

```bash
# Установка protoc и необходимых плагинов
sudo apt-get update
sudo apt-get install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Добавление плагинов в PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Генерация кода из proto-файлов
chmod +x ./generate_proto.sh
./generate_proto.sh
```

Скрипт `generate_proto.sh` автоматически:
1. Проверяет наличие необходимых файлов Google API (`annotations.proto` и `http.proto`) и скачивает их, если они отсутствуют
2. Сохраняет эти файлы локально в директории `third_party/google/api/` для дальнейшего использования
3. Генерирует код для всех сервисов с правильными импортами
4. Создает необходимые выходные директории

## Локальный запуск

Для локального запуска API Gateway:

```bash
# Запуск API Gateway
make run-gateway
```

API Gateway будет доступен по адресу http://localhost:8080.

## Развертывание

- Kind + Helm-umbrella
- Локальный registry на порту 5000

## Развитие проекта

Подробные инструкции в `docs/`. Документация по API Gateway доступна в [docs/api-gateway.md](docs/api-gateway.md).

## Микросервисы

| Сервис | Порт | Описание |
|--------|------|----------|
| API Gateway | 8080 (HTTP) | REST API для внешних клиентов, WebSocket для обновлений в реальном времени |
| Auth Service | 9100 (gRPC) | Управление аутентификацией и авторизацией пользователей |
| Device Service | 9200 (gRPC), 9101 (HTTP metrics) | Управление устройствами умного дома, получение статусов, отправка команд |
| Voice Service | 9300 (gRPC), 9201 (HTTP metrics) | Обработка голосовых команд для умного дома, распознавание интентов |
| Web Frontend | 8081 (HTTP) | Пользовательский интерфейс на React | 