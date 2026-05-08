#!/usr/bin/env bash
set -euo pipefail

REGISTRY="${REGISTRY:-ghcr.io/di3z1e/neuralog}"
TAG="${TAG:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"

echo "→ Building collector  ${REGISTRY}-collector:${TAG}"
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --file docker/collector.Dockerfile \
  --tag "${REGISTRY}-collector:${TAG}" \
  --tag "${REGISTRY}-collector:latest" \
  "${@}" \
  .

echo "→ Building UI         ${REGISTRY}-ui:${TAG}"
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --file docker/ui.Dockerfile \
  --tag "${REGISTRY}-ui:${TAG}" \
  --tag "${REGISTRY}-ui:latest" \
  "${@}" \
  .

echo "✓ Done. TAG=${TAG}"
