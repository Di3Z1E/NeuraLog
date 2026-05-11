REGISTRY       := ghcr.io/di3z1e/neuralog
TAG            ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
IMG            := $(REGISTRY):$(TAG)
HELM_CHART     := helm/neuralog
NAMESPACE      ?= log-system
RELEASE        ?= neuralog

.PHONY: help build push test lint helm-lint dev dev-down deploy upgrade uninstall clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Build ──────────────────────────────────────────────────────────────────────
build: ## Build the neuralog Docker image (embeds UI)
	docker build -f docker/neuralog.Dockerfile -t $(IMG) .
	docker tag $(IMG) $(REGISTRY):latest

# ── Push ───────────────────────────────────────────────────────────────────────
push: build ## Build + push the image
	docker push $(IMG)
	docker push $(REGISTRY):latest

# ── Test & Lint ────────────────────────────────────────────────────────────────
test: ## Run Go tests
	cd collector && go test ./... -race -count=1 -coverprofile=coverage.out
	cd collector && go tool cover -func=coverage.out | tail -1

lint: ## Lint Go and TypeScript
	cd collector && go vet ./...
	cd ui && npm run lint

helm-lint: ## Lint the Helm chart
	helm lint $(HELM_CHART) \
	  --set image.tag=test \
	  --strict

# ── Local dev ──────────────────────────────────────────────────────────────────
dev: ## Start full stack with hot-reload (docker-compose.override.yml)
	docker compose -f docker-compose.yml -f docker-compose.override.yml up --build

dev-down: ## Tear down local dev stack
	docker compose -f docker-compose.yml -f docker-compose.override.yml down -v

# ── Helm deploy ────────────────────────────────────────────────────────────────
deploy: ## Install the Helm chart (first time)
	helm upgrade --install $(RELEASE) $(HELM_CHART) \
	  --namespace $(NAMESPACE) \
	  --create-namespace \
	  --set image.tag=$(TAG) \
	  --wait

upgrade: ## Upgrade an existing release
	helm upgrade $(RELEASE) $(HELM_CHART) \
	  --namespace $(NAMESPACE) \
	  --set image.tag=$(TAG) \
	  --reuse-values \
	  --wait

uninstall: ## Uninstall the Helm release (PVC is retained)
	helm uninstall $(RELEASE) --namespace $(NAMESPACE)

# ── Cleanup ────────────────────────────────────────────────────────────────────
clean: ## Remove build artifacts
	rm -f collector/coverage.out
	rm -rf ui/dist ui/node_modules/.vite
