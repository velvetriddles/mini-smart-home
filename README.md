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
    - `v1/` - версия API
      - `common/` - общие типы данных, используемые несколькими сервисами
- `proto_generated/` - сгенерированные Go-файлы для общих типов
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

Генерация кода осуществляется с помощью protoc и скрипта generate_proto.sh через цели в Makefile:

```bash
# Установка необходимых инструментов protobuf
make install-tools

# Генерация кода из proto-файлов
make proto
```

На Windows используйте:

```bash
make install-tools
make proto-win
```

Цель `install-tools` автоматически устанавливает:
- protoc-gen-go - для генерации Go кода
- protoc-gen-go-grpc - для gRPC серверов и клиентов
- protoc-gen-grpc-gateway - для REST API из gRPC
- protoc-gen-openapiv2 - для OpenAPI спецификации

Скрипт `generate_proto.sh` (или `generate_proto.ps1` для Windows) автоматически:
1. Проверяет наличие необходимых файлов Google API (`annotations.proto` и `http.proto`) и скачивает их, если они отсутствуют
2. Сохраняет эти файлы локально в директории `third_party/google/api/` для дальнейшего использования
3. Генерирует код для всех сервисов с правильными импортами
4. Создает необходимые выходные директории
5. Генерирует общие типы в директории `proto_generated/`

Такой подход позволяет:
- Разместить общие типы данных в одном месте для всех сервисов
- Избежать дублирования кода
- Генерировать OpenAPI с корректными перекрестными ссылками

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