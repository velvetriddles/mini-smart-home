# Структура монорепозитория Smart-Home

## Структура сервисов

### API Gateway (`services/api-gateway`)

```
services/api-gateway/
├── cmd/
│   └── gateway/
│       └── main.go          # Точка входа API Gateway
├── internal/
│   ├── middleware/
│   │   └── auth.go          # JWT-аутентификация
│   └── server/
│       ├── grpc.go          # gRPC-клиент
│       └── http.go          # HTTP/WS сервер
├── proto/                   # Сгенерированные proto-файлы
├── Dockerfile               # Multi-stage Dockerfile
└── go.mod                   # Go модуль
```

API Gateway предоставляет REST-API и WebSocket для внешних клиентов, транслируя запросы во внутренние gRPC-сервисы. Реализует JWT-аутентификацию и интеграцию с Kafka для уведомлений в реальном времени.

### Auth Service (`services/auth`)

```
services/auth/
├── cmd/
│   └── auth/
│       └── main.go          # Точка входа сервиса
├── internal/
│   ├── config/              # Конфигурация
│   ├── model/               # Модели данных
│   ├── repository/          # Доступ к данным
│   └── service/             # Бизнес-логика
├── proto/                   # Сгенерированные proto-файлы
└── Dockerfile
```

Auth Service отвечает за аутентификацию и авторизацию пользователей, генерацию и валидацию JWT-токенов.

```
SmatHomeVKR/
├── .github/
│   └── workflows/
│       └── ci.yaml           # CI/CD конфигурация GitHub Actions
├── docs/
│   ├── context.md            # Описание проекта
│   ├── api-gateway.md        # Документация по API Gateway
│   └── structure.md          # Этот файл с описанием структуры
├── infra/
│   ├── helm-charts/
│   │   └── smarthome/        # Helm-umbrella чарт
│   │       ├── charts/       # Подчарты для каждого сервиса
│   │       ├── Chart.yaml    # Описание чарта и зависимостей
│   │       └── values.yaml   # Значения по умолчанию
│   ├── kind-config.yaml      # Конфигурация для локального кластера
│   └── registry-config.yaml  # Настройки локального registry
├── libs/
│   └── go.mod                # Модуль для общих библиотек
├── proto/
│   ├── smarthome/
│   │   └── v1/
│   │       ├── common/       # Общие определения типов
│   │       │   └── types.proto  
│   │       ├── auth.proto    # Описание сервиса аутентификации
│   │       └── device.proto  # Описание сервиса устройств
│   ├── buf.gen.yaml          # Конфигурация для генерации кода
│   └── buf.yaml              # Buf-линтер и настройки
├── proto_generated/          # Общие сгенерированные protobuf файлы
│   └── smarthome/
│       └── v1/               # Сгенерированные файлы общих типов
├── services/
│   ├── api-gateway/
│   │   ├── Dockerfile        # Multi-stage сборка
│   │   ├── go.mod            # Модуль Go
│   │   └── main.go           # Точка входа
│   ├── auth/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── main.go
│   ├── device/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── main.go
│   ├── voice/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── main.go
│   └── web/
│       ├── Dockerfile
│       ├── go.mod
│       └── main.go
├── web/
│   ├── src/
│   │   └── App.tsx           # React компонент
│   ├── Dockerfile            # Multi-stage сборка для фронтенда
│   └── package.json          # Зависимости и скрипты
├── go.work                   # Go Workspaces конфигурация
├── Makefile                  # Цели для сборки, генерации и деплоя
├── generate_proto.sh         # Скрипт для генерации protobuf файлов
└── README.md                 # Общее описание проекта
```

## Описание файлов в каждой директории

### Корневая директория
- `go.work` - описывает все Go модули в монорепозитории
- `Makefile` - содержит цели: install-tools, proto, build, kind, deploy
- `generate_proto.sh` - скрипт для генерации protobuf кода
- `README.md` - общее описание проекта и его структуры

### Сервисы (services/)
Каждый сервис содержит:
- `go.mod` - модуль Go с зависимостями сервиса
- `main.go` - точка входа с минимальной инициализацией
- `Dockerfile` - multi-stage сборка (Go → Alpine)
- `proto/` - сгенерированные protobuf файлы специфичные для сервиса

### Прото-файлы (proto/)
- Контракты gRPC в формате protobuf
- Настройки для генерации Go кода, REST, OpenAPI
- Директория `common/` содержит общие типы, используемые несколькими сервисами

### Общие сгенерированные protobuf файлы (proto_generated/)
- Содержит сгенерированные файлы для общих типов из `proto/smarthome/v1/common/`
- Используется всеми сервисами для доступа к общим типам

### Инфраструктура (infra/)
- Kind кластер для локальной разработки
- Helm чарт для деплоя всего приложения
- Подчарты для каждого микросервиса

### Web-интерфейс (web/)
- React приложение (TypeScript)
- Dockerfile для сборки и деплоя фронтенда

### Библиотеки (libs/)
- Общий код, используемый несколькими сервисами

### CI/CD (.github/)
- Пайплайн: тесты → сборка → деплой в Kind 