# Auth Service

Auth Service - сервис аутентификации и авторизации для системы "Умный дом". Обеспечивает безопасный доступ к API системы через JWT-токены.

## Функциональность

- **Аутентификация пользователей**: Вход, выход, обновление токенов
- **JWT-токены**: Создание, валидация и отзыв токенов
- **Управление пользователями**: Хранение учетных данных в PostgreSQL
- **Отзыв токенов**: Хранение отозванных токенов в Redis

## API (gRPC)

Сервис предоставляет следующие gRPC-методы:

- **Login**: Аутентификация по имени пользователя и паролю
- **Logout**: Отзыв токена доступа
- **Refresh**: Обновление токенов
- **ValidateToken**: Валидация токена и получение информации о пользователе

## Технологии

- Язык: Go 1.22
- Транспорт: gRPC
- Хранилище пользователей: PostgreSQL
- Хранилище отозванных токенов: Redis
- JWT-библиотека: github.com/golang-jwt/jwt/v5
- HTTP health check: Chi router (github.com/go-chi/chi/v5)
- Метрики: Prometheus (github.com/prometheus/client_golang)

## Конфигурация

Сервис настраивается через переменные окружения:

- `GRPC_PORT`: Порт для gRPC сервера (по умолчанию: 50051)
- `HTTP_PORT`: Порт для HTTP health check (по умолчанию: 9090)
- `POSTGRES_DSN`: URI для подключения к PostgreSQL
- `REDIS_DSN`: URI для подключения к Redis
- `JWT_SECRET`: Секретный ключ для подписи JWT токенов
- `JWT_TTL`: Время жизни JWT токенов (формат Go duration, по умолчанию: 24h)

## Структура базы данных

### PostgreSQL

Таблица `users`:
- `id`: UUID - уникальный идентификатор пользователя
- `username`: TEXT - имя пользователя (уникальное)
- `email`: TEXT - email пользователя (уникальный)
- `pass_hash`: TEXT - хеш пароля (bcrypt)
- `roles`: TEXT[] - массив ролей пользователя
- `created_at`: TIMESTAMP - время создания
- `updated_at`: TIMESTAMP - время обновления

### Redis

- `revoked:{jti}`: Хранит отозванные токены с TTL равным сроку истечения токена

## Локальный запуск

Для запуска сервиса локально используйте:

```bash
make run-auth
```

По умолчанию сервис запускается на портах:
- gRPC: 50051
- HTTP health check: 9090

## Тестирование

Для запуска всех тестов:

```bash
make test-auth
```

Для тестирования только функции генерации JWT:

```bash
make test-jwt
```

## Docker-образ

Сборка Docker-образа:

```bash
docker build -t smarthome/auth-service .
```

Запуск в Docker:

```bash
docker run -p 50051:50051 -p 9090:9090 \
  -e POSTGRES_DSN="postgres://user:pass@postgres:5432/smarthome" \
  -e REDIS_DSN="redis://redis:6379/0" \
  -e JWT_SECRET="your-secret-key" \
  smarthome/auth-service
```

## Health Check

HTTP health check доступен по адресу: `/healthz`
Prometheus метрики доступны по адресу: `/metrics` 