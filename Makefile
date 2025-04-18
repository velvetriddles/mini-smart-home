.PHONY: proto build kind deploy clean

# Глобальные переменные
REGISTRY ?= localhost:5000
VERSION ?= latest
SERVICES := api-gateway auth device voice web

# Proto-генерация
proto:
	@echo "Generating protobuf code..."
	cd proto && buf generate

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

# Очистка
clean:
	@echo "Cleaning up..."
	kind delete cluster --name smarthome-cluster
	docker rmi $(docker images | grep $(REGISTRY)/smarthome- | awk '{print $3}') || true 