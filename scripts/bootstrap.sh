#!/bin/bash
set -e

echo "==== Установка инструментов для Smart-Home ===="

# Проверка наличия curl
if ! command -v curl &> /dev/null; then
    echo "Требуется curl. Установите его перед запуском скрипта."
    exit 1
fi

# Проверка наличия docker
if ! command -v docker &> /dev/null; then
    echo "Требуется docker. Установите его перед запуском скрипта."
    exit 1
fi

# Установка buf
echo "Установка buf..."
BUF_VERSION="1.28.1"
curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" -o buf
chmod +x buf
sudo mv buf /usr/local/bin/buf
echo "buf установлен: $(buf --version)"

# Установка kind
echo "Установка kind..."
KIND_VERSION="0.20.0"
curl -sSLo ./kind "https://kind.sigs.k8s.io/dl/v${KIND_VERSION}/kind-$(uname)-amd64"
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
echo "kind установлен: $(kind --version)"

# Установка helm
echo "Установка helm..."
HELM_VERSION="3.13.3"
curl -sSL https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz | tar xz
sudo mv linux-amd64/helm /usr/local/bin/helm
rm -rf linux-amd64
echo "helm установлен: $(helm version --short)"

# Проверка kubectl
if ! command -v kubectl &> /dev/null; then
    echo "Рекомендуется установить kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl/"
else
    echo "kubectl уже установлен: $(kubectl version --client --short)"
fi

echo "==== Установка завершена ===="
echo "Теперь вы можете запустить проект Smart-Home:"
echo "1. make kind      # Создание кластера kind"
echo "2. make proto     # Генерация кода из proto-файлов"
echo "3. make build     # Сборка всех сервисов"
echo "4. make deploy    # Деплой в кластер" 