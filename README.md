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
- `proto/` - контракты gRPC
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

## Развертывание

- Kind + Helm-umbrella
- Локальный registry на порту 5000

## Развитие проекта

Подробные инструкции в `docs/`. 