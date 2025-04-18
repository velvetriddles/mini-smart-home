.PHONY: proto build kind deploy clean run-gateway clean-proto proto-win

# Глобальные переменные
REGISTRY ?= localhost:5000
VERSION ?= latest
SERVICES := api-gateway auth device voice

# Proto-генерация (Linux/Mac)
proto: clean-proto
	@echo "Generating protobuf code..."
	./generate_proto.sh

# Proto-генерация (Windows)
proto-win: clean-proto-win
	@echo "Generating protobuf code on Windows..."
	powershell -ExecutionPolicy Bypass -File ./generate_proto.ps1

# Очистка сгенерированных proto файлов (Linux/Mac)
clean-proto:
	@echo "Cleaning generated proto files..."
	rm -rf services/api-gateway/proto/* || true
	rm -rf services/auth/proto/* || true
	rm -rf services/device/proto/* || true
	rm -rf services/voice/proto/* || true
	mkdir -p services/api-gateway/proto/openapiv2
	mkdir -p services/auth/proto
	mkdir -p services/device/proto
	mkdir -p services/voice/proto

# Очистка сгенерированных proto файлов (Windows)
clean-proto-win:
	@echo "Cleaning generated proto files on Windows..."
	powershell -ExecutionPolicy Bypass -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue services/api-gateway/proto/*"
	powershell -ExecutionPolicy Bypass -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue services/auth/proto/*"
	powershell -ExecutionPolicy Bypass -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue services/device/proto/*"
	powershell -ExecutionPolicy Bypass -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue services/voice/proto/*"
	powershell -ExecutionPolicy Bypass -Command "New-Item -Path services/api-gateway/proto/openapiv2 -ItemType Directory -Force"
	powershell -ExecutionPolicy Bypass -Command "New-Item -Path services/auth/proto -ItemType Directory -Force"
	powershell -ExecutionPolicy Bypass -Command "New-Item -Path services/device/proto -ItemType Directory -Force"
	powershell -ExecutionPolicy Bypass -Command "New-Item -Path services/voice/proto -ItemType Directory -Force"

# Сборка всех сервисов
build:
	@echo "Building all services..."
	$(foreach service,$(SERVICES),$(MAKE) build-service SERVICE=$(service);)

# Сборка конкретного сервиса
build-service:
	@echo "Building $(SERVICE) service..."
	docker build -t $(REGISTRY)/smarthome-$(SERVICE):$(VERSION) -f services/$(SERVICE)/Dockerfile services/$(SERVICE)
	docker push $(REGISTRY)/smarthome-$(SERVICE):$(VERSION)

# Создание или запуск Kind кластера
kind:
	@echo "Setting up kind cluster..."
	kind create cluster --name smarthome-cluster --config infra/kind-config.yaml || true
	kubectl apply -f infra/registry-config.yaml

# Деплой в Kind через Helm
deploy:
	@echo "Deploying to kind cluster..."
	helm upgrade --install smarthome infra/helm-charts/smarthome --values infra/helm-charts/smarthome/values.yaml

# Запуск API Gateway локально
run-gateway:
	@echo "Running API Gateway locally..."
	cd services/api-gateway && go run cmd/gateway/main.go

# Очистка
clean:
	@echo "Cleaning up..."
	kind delete cluster --name smarthome-cluster
	docker rmi $(docker images | grep $(REGISTRY)/smarthome- | awk '{print $3}') || true 