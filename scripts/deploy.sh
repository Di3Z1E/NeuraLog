#!/usr/bin/env bash
set -euo pipefail

RELEASE="${RELEASE:-neuralog}"
NAMESPACE="${NAMESPACE:-log-system}"
TAG="${TAG:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
CHART="helm/neuralog"
VALUES_FILE="${VALUES_FILE:-}"

EXTRA_ARGS=()
if [[ -n "${VALUES_FILE}" ]]; then
  EXTRA_ARGS+=(-f "${VALUES_FILE}")
fi

echo "→ Deploying ${RELEASE} to namespace ${NAMESPACE} (tag: ${TAG})"

helm upgrade --install "${RELEASE}" "${CHART}" \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --set "image.collector.tag=${TAG}" \
  --set "image.ui.tag=${TAG}" \
  "${EXTRA_ARGS[@]}" \
  --wait \
  --timeout 5m

echo ""
helm get notes "${RELEASE}" -n "${NAMESPACE}"
