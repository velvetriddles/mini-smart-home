# Структура монорепозитория Smart-Home

```
SmatHomeVKR/
├── .github/
│   └── workflows/
│       └── ci.yaml           # CI/CD конфигурация GitHub Actions
├── docs/
│   ├── context.md            # Описание проекта
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
│   │       ├── auth.proto    # Описание сервиса аутентификации
│   │       └── device.proto  # Описание сервиса устройств
│   ├── buf.gen.yaml          # Конфигурация для генерации кода
│   └── buf.yaml              # Buf-линтер и настройки
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
└── README.md                 # Общее описание проекта
```

## Описание файлов в каждой директории

### Корневая директория
- `go.work` - описывает все Go модули в монорепозитории
- `Makefile` - содержит цели: proto, build, kind, deploy
- `README.md` - общее описание проекта и его структуры

### Сервисы (services/)
Каждый сервис содержит:
- `go.mod` - модуль Go с зависимостями сервиса
- `main.go` - точка входа с минимальной инициализацией
- `Dockerfile` - multi-stage сборка (Go → Alpine)

### Прото-файлы (proto/)
- Контракты gRPC в формате protobuf
- Настройки для генерации Go кода, REST, OpenAPI

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