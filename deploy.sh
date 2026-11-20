#!/bin/sh
set -eu

cd "$(dirname "$0")"
. ./.env

export IMAGE_TAG="${VERSION:?Укажите VERSION в .env}"

docker login "${IMAGE_REGISTRY:?Укажите IMAGE_REGISTRY}" \
  -u "${REGISTRY_USER:?Укажите REGISTRY_USER}" --password-stdin

docker compose pull
docker compose up -d --remove-orphans

curl -fsS --max-time 5 --retry 10 --retry-delay 2 --retry-all-errors \
  "http://127.0.0.1:${FRONTEND_PORT:-8888}/health" > /dev/null

printf '%s\n' "$VERSION" > version
docker image prune -f