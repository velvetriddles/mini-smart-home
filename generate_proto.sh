#!/bin/bash

PROTO_DIR=./proto
THIRD_PARTY_DIR=./third_party
OUT_DIR=./proto_generated
OPENAPI_DIR=./api-docs

# Определяем пути для Google API
GOOGLE_API_DIR="$THIRD_PARTY_DIR/google/api"

# Явно указываем пути к инструментам протобаф
PROTOC_GEN_GO="$(go env GOPATH)/bin/protoc-gen-go"
PROTOC_GEN_GO_GRPC="$(go env GOPATH)/bin/protoc-gen-go-grpc"
PROTOC_GEN_GRPC_GATEWAY="$(go env GOPATH)/bin/protoc-gen-grpc-gateway"
PROTOC_GEN_OPENAPIV2="$(go env GOPATH)/bin/protoc-gen-openapiv2"

# Проверяем, есть ли уже необходимые файлы Google API
if [ ! -f "$GOOGLE_API_DIR/annotations.proto" ] || [ ! -f "$GOOGLE_API_DIR/http.proto" ]; then
  echo "Локальная копия необходимых Google API файлов не найдена"
  
  # Создаем директории
  mkdir -p "$GOOGLE_API_DIR"
  
  # Скачиваем только необходимые файлы (без git clone всего репозитория)
  echo "Скачиваем необходимые файлы Google API..."
  curl -s https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -o "$GOOGLE_API_DIR/annotations.proto"
  curl -s https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -o "$GOOGLE_API_DIR/http.proto"
  
  echo "Файлы Google API успешно скачаны в $GOOGLE_API_DIR"
fi

# Создание директорий для выходных файлов
mkdir -p $OPENAPI_DIR
mkdir -p $OUT_DIR/smarthome/v1

# Общий набор опций
BASE="--proto_path=$PROTO_DIR --proto_path=$THIRD_PARTY_DIR --plugin=protoc-gen-go=$PROTOC_GEN_GO --plugin=protoc-gen-go-grpc=$PROTOC_GEN_GO_GRPC --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false"

# Очистка старых сгенерированных файлов из папки proto
echo "Удаление старых сгенерированных файлов из папки proto..."
find $PROTO_DIR -name "*.pb.go" -type f -delete
find $PROTO_DIR -name "*_grpc.pb.go" -type f -delete

# Очистка старых сгенерированных файлов в proto_generated
echo "Удаление старых сгенерированных файлов из папки proto_generated..."
rm -rf $OUT_DIR/smarthome

# Общие типы данных
echo "Генерация кода для общих типов данных..."
protoc $BASE \
  --go_out=$OUT_DIR --go-grpc_out=$OUT_DIR \
  $PROTO_DIR/smarthome/v1/common.proto

# Gateway: gRPC‑Gateway + OpenAPI + stubs
echo "Генерация кода для всех сервисов..."
protoc $BASE \
  --go_out=$OUT_DIR --go-grpc_out=$OUT_DIR \
  --plugin=protoc-gen-grpc-gateway=$PROTOC_GEN_GRPC_GATEWAY --grpc-gateway_out=$OUT_DIR --grpc-gateway_opt=paths=source_relative \
  $PROTO_DIR/smarthome/v1/*.proto

# OpenAPI документация для всех сервисов
echo "Генерация OpenAPI документации..."
protoc $BASE \
  --plugin=protoc-gen-openapiv2=$PROTOC_GEN_OPENAPIV2 --openapiv2_out=$OPENAPI_DIR \
  $PROTO_DIR/smarthome/v1/*.proto

echo "Генерация завершена успешно!"
echo "Все сервисы теперь должны импортировать proto из пакета github.com/velvetriddles/mini-smart-home/proto_generated/smarthome/v1"