#!/bin/sh
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ -f "$PROJECT_ROOT/.env" ]; then
  set -a
  . "$PROJECT_ROOT/.env"
  set +a
fi

VERSION="${VERSION:-latest}"
IMAGE="${NEXUS_REGISTRY}/${NEXUS_BACKEND_REPOSITORY}/sandbox-backend"

echo "Вход в registry ${NEXUS_REGISTRY}..."
printf '%s' "$NEXUS_PASSWORD" | docker login "$NEXUS_REGISTRY" -u "$NEXUS_USERNAME" --password-stdin

echo "Сборка ${IMAGE}:${VERSION}..."
docker build --target runtime -t "${IMAGE}:${VERSION}" -t "${IMAGE}:latest" "$SCRIPT_DIR"

echo "Публикация в registry..."
docker push "${IMAGE}:${VERSION}"
if [ "$VERSION" != "latest" ]; then
  docker push "${IMAGE}:latest"
fi

echo "Готово: ${IMAGE}:${VERSION}"
