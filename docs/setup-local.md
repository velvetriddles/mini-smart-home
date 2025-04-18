# Локальная настройка Smart-Home

Инструкция по настройке и запуску проекта Smart-Home в локальном окружении.

## Подготовка окружения

1. Установите необходимые инструменты, используя скрипт bootstrap:

```bash
chmod +x scripts/bootstrap.sh
./scripts/bootstrap.sh
```

Скрипт установит:
- protoc - для генерации кода из proto-файлов
- kind - для создания локального Kubernetes кластера
- helm - для управления Kubernetes-манифестами

2. Убедитесь, что у вас установлены:
   - Docker
   - kubectl
   - Go 1.22 или выше
   - make
   - curl - для скачивания Google API proto-файлов

## Запуск локального кластера

1. Создайте локальный Kubernetes кластер с помощью Kind:

```bash
make kind
```

Эта команда создаст кластер с тремя нодами (1 control-plane, 2 worker) и настроит маппинг портов:
- 8080 (для API Gateway)
- 8081 (для Web-интерфейса)

2. Проверьте, что кластер запущен:

```bash
kubectl cluster-info
```

## Сборка и деплой сервисов

1. Сгенерируйте код из proto-файлов:

```bash
# Установка protoc плагинов
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Добавление плагинов в PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Генерация кода
chmod +x ./generate_proto.sh
./generate_proto.sh
```

Скрипт `generate_proto.sh` автоматически:
- Проверяет наличие и при необходимости скачивает необходимые Google API proto-файлы
- Сохраняет их локально в `third_party/google/api/` для повторного использования
- Генерирует код для всех сервисов (api-gateway, auth, device)
- Создает необходимые выходные директории

Все необходимые файлы сохраняются локально, что убирает зависимость от внешних ресурсов при последующих запусках.

2. Соберите все сервисы и отправьте их в локальный registry:

```bash
make build
```

3. Разверните все компоненты в кластере:

```bash
make deploy
```

4. Проверьте статус деплоя:

```bash
kubectl get pods
```

## Доступ к сервисам

После успешного деплоя, вы можете получить доступ к сервисам:

1. API Gateway доступен по адресу: http://localhost:8080
2. Web-интерфейс доступен по адресу: http://localhost:8081

## Локальный запуск API Gateway

Для отладки можно запустить API Gateway локально без развертывания в Kubernetes:

```bash
# Запуск API Gateway
make run-gateway

# Запуск с указанием порта и адресов микросервисов
cd services/api-gateway && go run cmd/gateway/main.go --port=8082 --auth-service=localhost:50051
```

Локальный экземпляр будет доступен по адресу http://localhost:8080 (или указанному вами порту).

Для тестирования WebSocket-соединения можно использовать клиенты вроде [websocat](https://github.com/vi/websocat):

```bash
# Подключение к WebSocket
websocat ws://localhost:8080/ws

# Или с помощью curl для REST API
curl http://localhost:8080/health
```

## Тестовые учетные данные

Для доступа к системе используйте следующие учетные данные:

- Логин: admin
- Пароль: admin123

## Мониторинг и отладка

1. Просмотр логов сервисов:

```bash
kubectl logs -l app=api-gateway
kubectl logs -l app=auth
kubectl logs -l app=device
kubectl logs -l app=voice
```

2. Доступ к UI подов:

```bash
kubectl port-forward svc/api-gateway 8080:8080
```

## Остановка и очистка

Для остановки и удаления всех ресурсов выполните:

```bash
make clean
```

Эта команда удалит кластер Kind и все созданные контейнеры. 