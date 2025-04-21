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
│   ├── api-gateway.md        # Документация по API Gateway
│   ├── architecture-c4.md    # Схемы архитектуры в формате C4
│   ├── context.md            # Описание проекта
│   ├── flow-sequence.md      # Диаграммы последовательности процессов
│   ├── glossary.md           # Глоссарий проекта
│   ├── setup-local.md        # Руководство по локальной настройке
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
├── scripts/
│   └── bootstrap.sh          # Скрипт инициализации окружения
├── services/
│   ├── api-gateway/          # Сервис API Gateway
│   │   ├── Dockerfile        # Multi-stage сборка
│   │   ├── go.mod            # Модуль Go
│   │   └── main.go           # Точка входа
│   ├── auth/                 # Сервис аутентификации
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── main.go
│   ├── device/               # Сервис управления устройствами
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── main.go
│   └── voice/                # Сервис голосового управления
│       ├── Dockerfile
│       ├── go.mod
│       └── main.go
├── third_party/              # Сторонние зависимости
│   └── google/               # Google API и прото-файлы
├── web/                      # Веб-интерфейс
│   ├── src/
│   │   └── App.tsx           # React компонент
│   ├── public/               # Статические файлы
│   ├── node_modules/         # Зависимости NPM
│   ├── Dockerfile            # Multi-stage сборка для фронтенда
│   ├── INSTALL.md            # Инструкция по установке
│   ├── README.md             # Описание фронтенда
│   ├── index.html            # Входная точка HTML
│   ├── package.json          # Зависимости и скрипты
│   ├── tailwind.config.js    # Конфигурация Tailwind CSS
│   ├── tsconfig.json         # Настройки TypeScript
│   └── vite.config.ts        # Конфигурация Vite
├── docker-compose.yml        # Конфигурация Docker Compose
├── go.work                   # Go Workspaces конфигурация
├── go.work.sum               # Контрольные суммы зависимостей
├── Makefile                  # Цели для сборки, генерации и деплоя
├── TODO.md                   # Список задач
├── generate_proto.sh         # Скрипт для генерации protobuf файлов
└── README.md                 # Общее описание проекта
```

## Описание файлов в каждой директории

### Корневая директория
- `go.mod` - единый модуль Go для всего проекта (github.com/velvetriddles/mini-smart-home)
- `Makefile` - содержит цели: install-tools, proto, build, kind, deploy
- `docker-compose.yml` - конфигурация для запуска сервисов через Docker Compose
- `generate_proto.sh` - скрипт для генерации protobuf кода
- `README.md` - общее описание проекта и его структуры
- `TODO.md` - список запланированных задач и улучшений

### Сервисы (services/)
Каждый сервис содержит:
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
- React приложение на TypeScript
- Vite для сборки и разработки
- Tailwind CSS для стилизации
- Dockerfile для сборки и деплоя фронтенда

### Библиотеки (libs/)
- Общий код, используемый несколькими сервисами

### Скрипты (scripts/)
- `bootstrap.sh` - скрипт для быстрой инициализации окружения разработки

### Сторонние зависимости (third_party/)
- Внешние зависимости, необходимые для сборки проекта
- Google API и protobuf определения

### Документация (docs/)
- Полная документация по архитектуре, API и настройке
- Диаграммы и схемы архитектуры в формате C4
- Руководство по локальной разработке

### CI/CD (.github/)
- Пайплайн: тесты → сборка → деплой в Kind 