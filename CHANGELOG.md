# Changelog

All notable changes to this project are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## [v0.2.1] — 2026-05-12

### Fixed

- **Full pod log history on attach** — the collector was capping log streams at `SinceSeconds=60`, so only the last minute of logs was ever fetched when a pod was first attached. First connect now requests all logs from container start. Reconnects resume from the last observed timestamp (`SinceTime + 1ns`) to prevent both gaps and duplicate lines across retries.
- **Sidebar showing no pods** — `hasLogs` was computed once at pod-discovery time and cached in memory. Because no log file exists at the instant a pod is first seen by the informer, the flag was always `false` for live pods. The sidebar filters to `hasLogs=true` only, so no pods ever appeared even though logs were actively being written. `ListPods` now does a live `os.Stat` check per pod on every call.

### Changed — Helm

- **Removed chart-managed namespace** — `namespace.yaml` was creating the namespace as a Helm-managed resource, causing a chicken-and-egg failure on every fresh install (resources couldn't be deployed into a namespace that didn't exist yet). `--create-namespace` on the CLI now handles it, which is the standard Helm pattern and a requirement for Helm Hub publishing.
- **`storageClassName` defaults to cluster default** — was hardcoded to `local-path` (k3s-specific). Now defaults to `""` so the PVC template omits the field entirely and Kubernetes picks the cluster's default StorageClass. Override it in your values file if needed.
- **Renamed `k3s-values.yaml` → `values-local.yaml`** — cluster-specific override files follow a `values-<env>.yaml` convention. The file now contains only the overrides that differ from defaults.
- **`make deploy` is idempotent on any cluster** — uses `helm upgrade --install --create-namespace` and works on first install and every subsequent upgrade. No cluster-specific make targets.

---

## [v0.2.0] — 2026-05-11

### Added

- **Single self-contained image** — the React UI is now embedded into the Go binary at build time via `//go:embed`. The collector serves both the REST/WebSocket API and the SPA from a single distroless container. The previous two-image setup (collector + nginx) is gone.
  - 3-stage Docker build: `node:20-alpine` → `golang:1.25-alpine` → `distroless/static-debian12:nonroot`
  - `/assets/*` served with immutable cache headers
  - SPA catch-all falls back to `/` (not `/index.html`) to avoid `http.FileServer`'s built-in 301 redirect
  - Security headers middleware added to all responses

- **Unit test suite — 22 tests** covering:
  - `redactor` — builtin patterns, custom patterns, reload, disabled mode
  - `config` — defaults, environment variable seeding, persist/load round-trip
  - `store` — Append, ReadLines (search/level/time filters), rotation, ListPods, DiskUsageBytes, parseLineTime
  - `janitor` — retention enforcement, rotated file cleanup, empty-dir pruning

### Removed

- Legacy proof-of-concept Kubernetes manifests (`00-namespace.yaml`, `01-storage.yaml`, `02-rbac.yaml`, `03-deployment.yaml`, `04-retention-job.yaml`) — superseded by the Helm chart.
- `scripts/build.sh` and `scripts/deploy.sh` — duplicated Makefile targets.
- `docker/collector.Dockerfile`, `docker/ui.Dockerfile`, `docker/nginx.conf` — replaced by the single `docker/neuralog.Dockerfile`.

### Fixed

- **Security: `golang.org/x/net` upgraded to v0.53.0** — resolves GO-2026-4918 (HTTP/2 infinite loop). Requires Go 1.25+; `go.mod` and the builder image updated accordingly.
- **Security: Gitleaks false positives on test fixtures** — `redactor_test.go` contains intentional secret-shaped strings as test input (AWS key ID, GitHub token, JWT). Added `.gitleaks.toml` allowlist scoped to the test file path.

---

## [v0.1.0] — 2026-05-11

### Added

- **Go collector** (`collector/`) — Kubernetes log aggregation service with two subcommands: `serve` and `janitor`.
  - Pod discovery via `SharedInformerFactory` (5-minute resync). Attaches to all containers (init + regular) in Running pods across all non-excluded namespaces.
  - Per-container log streaming with automatic retry on stream failure.
  - Flat-file log store: one `.log` file per pod at `<basePath>/<namespace>/<pod>.log`. Supports size-based log rotation and configurable keep count.
  - WebSocket fan-out hub: single-goroutine channel-based broadcast, 4096-message buffer, per-client 256-message send buffer, slow-consumer eviction.
  - Secret redaction pipeline with 11 built-in regex patterns: JWT, Bearer tokens, AWS key ID, AWS secret key, generic API keys, passwords, generic secrets, database URLs, basic-auth URLs, PEM private keys, credit card numbers. Custom patterns supported.
  - Runtime config system: settings persisted to `.neuralog.json` on the log volume. All changes (redaction rules, namespace exclusions, rotation policy, retention) take effect immediately without a pod restart.
  - Storage quota watcher: background goroutine evicts oldest log files when total disk usage exceeds the configured limit.
  - Janitor: deletes log files older than `retentionDays`, prunes empty namespace directories. Designed to run as a Kubernetes CronJob.

- **HTTP/WebSocket API** (`/api/v1`):
  - `GET /healthz` — liveness probe
  - `GET /ws?namespace=&pod=` — live log tail over WebSocket; seeds the last 200 lines on connect then streams live
  - `GET /api/v1/pods` — list all known pods with status and `hasLogs` flag; merges live (in-memory) and historical (on-disk) pods
  - `GET /api/v1/logs/{namespace}/{pod}` — fetch stored logs with optional `search`, `level`, `from`, `to` (RFC3339), and `lines` filters
  - `GET /api/v1/download/{namespace}/{pod}` — download raw log file
  - `GET /api/v1/config` — retrieve current runtime config including live disk usage
  - `PUT /api/v1/config` — update runtime config; hot-reloads redactor and applies namespace exclusions immediately

- **React UI** (`ui/`) — dark-theme single-page application built with React 18, TypeScript 5, and Vite 5.
  - Sidebar: pods grouped and sorted by namespace, filtered to pods with logs on disk, collapsible namespaces, inline search filter.
  - Log viewer: virtualized list via TanStack Virtual (handles millions of lines). Dual mode — live (WebSocket, client-side filter) and history (REST, server-side filter). Auto-tail snaps to bottom when scrolled within 60px of the end; "Jump to latest" button when scrolled away.
  - TopBar: live/history toggle, full-text search, log-level filter (ALL / TRACE / DEBUG / INFO / WARN / ERROR / FATAL), date-range picker, download link.
  - Settings modal: three tabs — Storage (quota, rotation), Collection (namespace exclusions), Redaction (toggle + custom pattern table). All changes applied live via `PUT /api/v1/config`.
  - WebSocket hook with exponential backoff reconnect (1s → 30s cap).
  - State managed with Zustand; server state with TanStack Query.

- **Helm chart** (`helm/neuralog/`) — production-ready chart for any Kubernetes distribution.
  - ClusterRole/ClusterRoleBinding granting `get`, `list`, `watch` on pods and `pods/log`.
  - PersistentVolumeClaim for log storage (configurable size, StorageClass, access mode).
  - Deployment with `Recreate` strategy (required for `ReadWriteOnce` PVC), security context (`runAsNonRoot`, read-only root filesystem, dropped capabilities), liveness and readiness probes.
  - CronJob running `neuralog janitor` on a configurable schedule (default: daily at 02:00 UTC).
  - NetworkPolicy: restricts ingress to the service port and adds an optional egress rule for NFS servers.
  - Ingress, HPA, and ServiceAccount templates — all optional and disabled by default.

- **CI/CD** (`.github/workflows/`):
  - `ci-collector.yml` — Go build, vet, test with race detector, govulncheck
  - `ci-ui.yml` — TypeScript typecheck, ESLint, Vite production build
  - `ci-helm.yml` — `helm lint --strict`, template render, ingress variant
  - `ci-docker.yml` — PR-only full Docker build validation (no push)
  - `security.yml` — govulncheck, Trivy FS/IaC SARIF, Gitleaks full-history scan
  - `codeql.yml` — CodeQL SAST for Go and TypeScript
  - `release.yml` — tag-triggered: build and push image to GHCR, create GitHub Release
  - `helm-publish.yml` — chart-releaser publishes chart package to `gh-pages` on release

### Fixed

- **Helm: `nfs.server` guarded in NetworkPolicy** — missing default caused `helm lint --strict` to fail with a nil pointer when `networkPolicy.enabled=true`. Added `nfs.server: ""` default and wrapped the NFS egress block in a conditional.
- **CI: all pipeline failures resolved** — pinned Helm to v3.14.4, switched Trivy to apt install, added missing `.eslintrc.cjs`.

---

[v0.2.1]: https://github.com/Di3Z1E/NeuraLog/releases/tag/v0.2.1
[v0.2.0]: https://github.com/Di3Z1E/NeuraLog/releases/tag/v0.2.0
[v0.1.0]: https://github.com/Di3Z1E/NeuraLog/releases/tag/v0.1.0
