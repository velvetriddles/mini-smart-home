# Device Service

Сервис управления устройствами умного дома в системе Smart-Home. Предоставляет gRPC API для управления устройствами, получения их статусов и настройки.

## Функциональность

- Получение информации об устройствах
- Фильтрация устройств по типу и статусу
- Управление устройствами (включение/выключение, настройка параметров)
- Потоковая передача обновлений статуса устройств

## Технический стек

- Go 1.22
- gRPC/Protocol Buffers
- In-memory хранилище (в первой итерации)
- Prometheus для метрик
- gRPC Health Checking Protocol

## Конфигурация

Device Service может быть настроен с помощью переменных окружения:

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `PORT` | gRPC порт сервиса | `9200` |
| `HTTP_PORT` | HTTP порт для метрик и health-check | `9101` |
| `LOG_LEVEL` | Уровень логирования (debug, info, warn, error) | `info` |
| `METRICS_ENABLED` | Включение/выключение Prometheus метрик | `true` |

## Локальный запуск

### Через Make

```bash
# Запуск сервиса
make run-device

# Запуск тестов
make test-device

# Сборка Docker-образа
make docker-build

# Запуск Docker-контейнера
make docker-run
```

### Через Go напрямую

```bash
# Запуск сервиса
cd services/device
go run cmd/device/main.go --port=9200 --http-port=9101
```

## Примеры использования с grpcurl

### Получение списка устройств

```bash
grpcurl -plaintext localhost:9200 smarthome.v1.DeviceService/ListDevices
```

### Получение устройства по ID

```bash
grpcurl -plaintext -d '{"id": "device-id-here"}' localhost:9200 smarthome.v1.DeviceService/GetDevice
```

### Фильтрация устройств по типу

```bash
grpcurl -plaintext -d '{"type": "lamp", "online_only": true}' localhost:9200 smarthome.v1.DeviceService/ListDevices
```

### Управление устройством

```bash
grpcurl -plaintext -d '{
  "id": "device-id-here",
  "command": {
    "action": "turn_on",
    "parameters": {
      "level": "80"
    }
  }
}' localhost:9200 smarthome.v1.DeviceService/ControlDevice
```

### Потоковая передача статусов устройств

```bash
grpcurl -plaintext -d '{"subscribe_all": true}' localhost:9200 smarthome.v1.DeviceService/StreamStatuses
```

## Структура каталогов

```
services/device/
├── cmd/
│   └── device/
│       └── main.go          # Точка входа
├── internal/
│   ├── server/
│   │   └── grpc.go          # gRPC-сервер с методами
│   ├── datastore/
│   │   ├── memory.go        # In-memory хранилище
│   │   └── memory_test.go   # Тесты хранилища
│   └── model/
│       └── device.go        # Модели данных
├── proto/                   # Сгенерированные proto-файлы
├── sql/
│   └── schema.sql           # Схема БД для будущей миграции на PostgreSQL
├── Dockerfile               # Multi-stage Dockerfile
├── go.mod                   # Зависимости Go-модуля
└── README-device.md         # Этот файл
```

## Health-check

Device Service реализует gRPC Health Checking Protocol, что позволяет использовать его с Kubernetes для проверки работоспособности.

```bash
# Проверка gRPC health
grpcurl -plaintext localhost:9200 grpc.health.v1.Health/Check

# Проверка HTTP health
curl http://localhost:9101/health
```

## Метрики Prometheus

Метрики доступны по адресу http://localhost:9101/metrics и включают стандартные метрики Go, а также специфичные для сервиса метрики.

## Тестирование

```bash
# Запуск всех тестов
cd services/device
go test ./...

# Запуск тестов хранилища
go test ./internal/datastore

# Запуск с флагом coverage
go test -cover ./...
``` 