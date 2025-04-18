# Voice Service

Сервис голосового управления для системы Smart-Home. Обрабатывает текстовые и голосовые команды, определяет интенты и выполняет соответствующие действия через другие сервисы.

## Функциональность

- Распознавание интентов в текстовых командах (NLU)
- Извлечение сущностей из команд (устройства, комнаты, значения)
- Выполнение действий через взаимодействие с другими сервисами (Device, Auth)
- Аутентификация пользователей через Auth Service
- Отладочный API для просмотра доступных интентов

## Технический стек

- Go 1.22
- gRPC/Protocol Buffers
- Prometheus для метрик
- gRPC Health Checking Protocol

## Конфигурация

Voice Service может быть настроен с помощью переменных окружения или флагов командной строки:

| Переменная/Флаг | Описание | Значение по умолчанию |
|-----------------|----------|----------------------|
| `PORT` | gRPC порт сервиса | `9300` |
| `HTTP_PORT` | HTTP порт для метрик и health-check | `9201` |
| `DEVICE_ADDR` | Адрес Device Service | `localhost:9200` |
| `AUTH_ADDR` | Адрес Auth Service | `localhost:9100` |
| `LOG_LEVEL` | Уровень логирования (debug, info, warn, error) | `info` |

## Локальный запуск

### Через Make

```bash
# Запуск сервиса
make run-voice

# Запуск тестов
make test-voice

# Сборка Docker-образа
make docker-build

# Запуск Docker-контейнера
make docker-run
```

### Через Go напрямую

```bash
# Запуск сервиса
cd services/voice
go run cmd/voice/main.go --port=9300 --http-port=9201
```

## Примеры использования с grpcurl

### Взаимодействие с сервисом через двунаправленный поток команд

Для взаимодействия с сервисом через `RecognizeCommand` необходимо использовать `-d @` и вводить JSON-объекты в интерактивном режиме:

```bash
grpcurl -plaintext -d @ localhost:9300 smarthome.v1.VoiceService/RecognizeCommand
{"session_id": "test-session-123", "text": "включи свет в гостиной"}
```

Важно: в реальном клиенте необходимо передавать в метаданных токен авторизации:

```bash
grpcurl -plaintext -H "authorization: Bearer YOUR_ACCESS_TOKEN" -d @ localhost:9300 smarthome.v1.VoiceService/RecognizeCommand
```

### Получение списка поддерживаемых интентов

```bash
grpcurl -plaintext localhost:9300 smarthome.v1.VoiceService/ListIntents
```

## Поддерживаемые интенты

| Интент | Описание | Пример команды |
|--------|----------|----------------|
| TurnOn | Включение устройства | "Включи свет в гостиной" |
| TurnOff | Выключение устройства | "Выключи лампу на кухне" |
| GetTemperature | Запрос температуры | "Какая температура в спальне" |
| SetTemperature | Установка температуры | "Установи температуру 22 градуса в спальне" |

## Структура каталогов

```
services/voice/
├── cmd/
│   └── voice/
│       └── main.go          # Точка входа
├── internal/
│   ├── server/
│   │   ├── grpc.go          # gRPC-сервер с методами
│   │   └── grpc_test.go     # Тесты сервера
│   ├── nlu/
│   │   ├── simple.go        # Простая NLU на основе ключевых слов
│   │   └── simple_test.go   # Тесты NLU
│   └── model/
│       └── intent.go        # Модели данных для интентов
├── proto/                   # Сгенерированные proto-файлы
├── Dockerfile               # Multi-stage Dockerfile
├── Makefile                 # Makefile с командами
├── go.mod                   # Зависимости Go-модуля
└── README-voice.md          # Этот файл
```

## Health-check

Voice Service реализует gRPC Health Checking Protocol, что позволяет использовать его с Kubernetes для проверки работоспособности.

```bash
# Проверка gRPC health
grpcurl -plaintext localhost:9300 grpc.health.v1.Health/Check

# Проверка HTTP health
curl http://localhost:9201/health
```

## Метрики Prometheus

Метрики доступны по адресу http://localhost:9201/metrics и включают стандартные метрики Go, а также специфичные для сервиса метрики.

## Ограничения текущей версии

- Поддерживаются только текстовые команды, распознавание аудио не реализовано
- Ограниченный набор интентов и сущностей
- Упрощенный алгоритм NLU на основе ключевых слов
- Отсутствие механизма обучения модели

## Планируемые улучшения

- Интеграция с внешними NLU сервисами (Dialogflow, Rasa)
- Добавление распознавания аудио (ASR)
- Расширение набора интентов и сущностей
- Персонализация команд на основе истории взаимодействия пользователя 