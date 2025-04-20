#!/bin/bash

PROTO_DIR=./proto
THIRD_PARTY_DIR=./third_party
OUT_GATEWAY=./services/api-gateway/proto
OUT_AUTH=./services/auth/proto
OUT_DEVICE=./services/device/proto
OUT_VOICE=./services/voice/proto
OUT_COMMON=./proto_generated

# Определяем пути для Google API
GOOGLE_API_DIR="$THIRD_PARTY_DIR/google/api"

# Определение пути к Google protobuf
PROTOC_INC="$(go env GOPATH)/pkg/mod/github.com/golang/protobuf@*/src"

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
mkdir -p $OUT_GATEWAY/openapiv2
mkdir -p $OUT_AUTH
mkdir -p $OUT_DEVICE
mkdir -p $OUT_VOICE
mkdir -p $OUT_COMMON

# Общий набор опций
BASE="--proto_path=$PROTO_DIR --proto_path=$THIRD_PARTY_DIR --proto_path=$PROTOC_INC --plugin=protoc-gen-go=$PROTOC_GEN_GO --plugin=protoc-gen-go-grpc=$PROTOC_GEN_GO_GRPC --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false"

# Общие типы данных
echo "Генерация кода для общих типов данных..."
protoc $BASE \
  --go_out=$OUT_COMMON --go-grpc_out=$OUT_COMMON \
  $PROTO_DIR/smarthome/v1/common/*.proto

# Gateway: gRPC‑Gateway + OpenAPI + stubs
echo "Генерация кода для API Gateway..."
protoc $BASE \
  --go_out=$OUT_GATEWAY --go-grpc_out=$OUT_GATEWAY \
  --plugin=protoc-gen-grpc-gateway=$PROTOC_GEN_GRPC_GATEWAY --grpc-gateway_out=$OUT_GATEWAY --grpc-gateway_opt=paths=source_relative \
  $PROTO_DIR/smarthome/v1/*.proto

# OpenAPI документация для всех сервисов
echo "Генерация OpenAPI документации..."
protoc $BASE \
  --plugin=protoc-gen-openapiv2=$PROTOC_GEN_OPENAPIV2 --openapiv2_out=$OUT_GATEWAY/openapiv2 \
  $PROTO_DIR/smarthome/v1/*.proto

# Auth‑service
echo "Генерация кода для Auth Service..."
protoc $BASE \
  --go_out=$OUT_AUTH --go-grpc_out=$OUT_AUTH \
  $PROTO_DIR/smarthome/v1/auth.proto

# Device‑service
echo "Генерация кода для Device Service..."
protoc $BASE \
  --go_out=$OUT_DEVICE --go-grpc_out=$OUT_DEVICE \
  $PROTO_DIR/smarthome/v1/device.proto

# Voice‑service
echo "Генерация кода для Voice Service..."
protoc $BASE \
  --go_out=$OUT_VOICE --go-grpc_out=$OUT_VOICE \
  $PROTO_DIR/smarthome/v1/voice.proto

echo "Генерация завершена успешно!"