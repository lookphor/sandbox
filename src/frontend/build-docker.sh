#!/usr/bin/env bash
set -euo pipefail

IMAGE="${IMAGE:-sandbox-frontend}"
TAG="${TAG:-latest}"
REGISTRY="${REGISTRY:-}"

FULL="${IMAGE}:${TAG}"
[ -n "${REGISTRY}" ] && FULL="${REGISTRY}/${FULL}"

cd "$(dirname "$0")"

docker build \
    --label "org.opencontainers.image.source=$(git config --get remote.origin.url 2>/dev/null || echo unknown)" \
    --label "org.opencontainers.image.revision=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" \
    --label "org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -t "${FULL}" \
    .

echo "✓ собрано: ${FULL}"