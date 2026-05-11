# NeuraLog

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Helm](https://img.shields.io/badge/Helm-3.x-0F1689?style=flat-square&logo=helm&logoColor=white)](helm/neuralog)
[![License](https://img.shields.io/github/license/Di3Z1E/NeuraLog?style=flat-square)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Di3Z1E/NeuraLog?style=flat-square)](https://github.com/Di3Z1E/NeuraLog/releases)

Kubernetes-native log aggregation with real-time streaming, sensitive-data redaction, and a web UI — packaged as a single self-contained binary.

---

## Overview

NeuraLog discovers running pods via the Kubernetes Informer API, streams their logs to persistent storage with automatic redaction, and serves a dark-theme web UI with live WebSocket streaming, historical search, and a settings panel for runtime configuration. All operational parameters can be changed from the UI without restarting the pod.

The Go binary embeds the compiled React frontend. A single `distroless/static:nonroot` container handles the API, WebSocket stream, and SPA — no sidecar, no nginx, no DaemonSet.

## Features

- **Real-time streaming** — WebSocket tail with 10k-line virtual scroll and exponential-backoff reconnect
- **Sensitive-data redaction** — JWT, Bearer tokens, AWS keys, passwords, database URLs, credit card numbers stripped before any write to disk; custom regex patterns configurable from the UI
- **Informer-based discovery** — pod events delivered in milliseconds via Kubernetes watch, no polling
- **Runtime configuration** — storage quota, log rotation, retention, namespace exclusions, and redaction rules all changeable from the UI; no pod restart required
- **Log rotation and quota** — per-pod file rotation by size with configurable history depth; hard storage cap with oldest-first eviction
- **Automated retention** — `janitor` subcommand runs as a nightly CronJob; TTL is configurable from the UI
- **Hardened defaults** — `distroless/static:nonroot`, `readOnlyRootFilesystem`, all capabilities dropped, NetworkPolicy included

---

## Quick Start

### Prerequisites

- Docker and `docker compose`
- `~/.kube/config` pointing at a live cluster
- Go 1.25+ to build from source
- Node 20+ for UI development

### Run locally

```bash
git clone https://github.com/Di3Z1E/NeuraLog
cd NeuraLog
make dev
```

| Service | URL |
|---------|-----|
| UI (Vite dev server) | http://localhost:3000 |
| API | http://localhost:8080 |

### Run without Docker

```bash
cd collector
go run ./cmd/neuralog serve
```

Environment variables seed the config on first boot. Once `.neuralog.json` exists on the storage volume, the UI-saved config takes precedence.

| Variable | Default | Description |
|----------|---------|-------------|
| `NEURALOG_LOG_BASE_PATH` | `/mnt/logs` | Base directory for log storage |
| `NEURALOG_LISTEN_ADDR` | `:8080` | HTTP listen address |
| `NEURALOG_EXCLUDE_NAMESPACES` | `log-system,kube-system` | Comma-separated namespaces to skip |
| `NEURALOG_REDACT_ENABLED` | `true` | Enable sensitive-data redaction |
| `NEURALOG_RETENTION_DAYS` | `7` | Log retention TTL in days |
| `KUBECONFIG` | _(in-cluster)_ | Path to kubeconfig for out-of-cluster use |

---

## Helm Deployment

```bash
helm upgrade --install neuralog helm/neuralog \
  --namespace log-system \
  --create-namespace \
  --set image.tag=v0.2.0 \
  --wait
```

Common overrides:

```yaml
# values.override.yaml
storage:
  storageClassName: nfs-client
  storageSize: 100Gi

nfs:
  server: "10.0.0.10"

ingress:
  enabled: true
  host: logs.yourdomain.com
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod

collector:
  excludeNamespaces: "log-system,kube-system,monitoring"

retention:
  days: 14
```

```bash
helm upgrade --install neuralog helm/neuralog \
  -f values.override.yaml \
  --namespace log-system --create-namespace
```

Full values reference: [`helm/neuralog/values.yaml`](helm/neuralog/values.yaml)

---

## Settings

Open the **gear icon** in the top-right of the UI to access runtime configuration. Changes are written to `.neuralog.json` on the storage volume and applied immediately.

### Storage

| Setting | Description |
|---------|-------------|
| Storage quota (GiB) | Hard cap on total disk usage. Exceeded quota evicts oldest files first. `0` means unlimited. |
| Rotation size (MB) | Rotate a pod's log file at this size. Produces `pod.log.1`, `pod.log.2`, etc. `0` disables rotation. |
| Rotated files to keep | Maximum number of rotated files to keep per pod. |
| Retention (days) | The nightly janitor deletes log files older than this many days. |

### Collection

Add or remove excluded namespaces. Newly excluded namespaces stop streaming immediately; newly included ones are picked up on the next informer event.

### Redaction

Toggle the master redaction switch and manage custom regex patterns on top of the built-in rules. Changes take effect on the next incoming log line.

---

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/pods` | List tracked pods — live and historical |
| `GET` | `/api/v1/logs/{namespace}/{pod}` | Historical logs (`?lines=N&search=S&level=L&from=T&to=T`) |
| `GET` | `/api/v1/download/{namespace}/{pod}` | Download raw log file |
| `GET` | `/api/v1/config` | Current runtime config (includes `storageUsedGB`) |
| `PUT` | `/api/v1/config` | Update config; hot-reloads redactor and namespace exclusions |
| `WS`  | `/ws?namespace=N&pod=P` | Live log stream; seeds last 200 lines on connect |
| `GET` | `/healthz` | Health check |

### Config schema

```json
{
  "storageQuotaGB":    10,
  "rotationMaxMB":     100,
  "rotationKeepFiles": 5,
  "retentionDays":     7,
  "excludeNamespaces": ["log-system", "kube-system"],
  "redactEnabled":     true,
  "customPatterns": [
    { "id": "my-rule", "pattern": "secret-[a-z0-9]+", "replace": "[REDACTED:CUSTOM]" }
  ]
}
```

---

## Redaction

All patterns run in the collection pipeline before any write to disk or broadcast over WebSocket.

| Pattern | Replacement |
|---------|-------------|
| JWT tokens | `[REDACTED:JWT]` |
| Bearer tokens | `[REDACTED:BEARER_TOKEN]` |
| AWS key IDs | `[REDACTED:AWS_KEY_ID]` |
| AWS secret keys | `[REDACTED:AWS_SECRET]` |
| Generic API keys | `[REDACTED:API_KEY]` |
| Passwords in log lines | `[REDACTED:PASSWORD]` |
| Generic secrets and tokens | `[REDACTED:SECRET]` |
| Database URLs with credentials | `[REDACTED:DB_URL]` |
| Basic auth in URLs | `[REDACTED:CREDENTIALS]` |
| Private key PEM blocks | `[REDACTED:PRIVATE_KEY]` |
| Credit card numbers | `[REDACTED:CARD]` |

Custom patterns can be added from the **Settings > Redaction** tab or via `PUT /api/v1/config`.

---

## Architecture

```
+----------------------------------------------+
|              Kubernetes Cluster              |
|                                              |
|  [Pod A]  [Pod B]  [Pod C]  ...              |
|      |        |        |                     |
|      +--------+--------+                     |
|               |  Informer watch              |
|               v                              |
|     +--------------------+                   |
|     |   NeuraLog Pod     |                   |
|     |                    |                   |
|     |  [neuralog binary] |                   |
|     |   - client-go      |  <- pod watch     |
|     |   - redact + write |                   |
|     |   - REST API       |  <- /api/v1/...   |
|     |   - WebSocket      |  <- /ws           |
|     |   - embedded UI    |  <- /             |
|     |       |            |                   |
|     |  [/mnt/logs]       |                   |
|     |   +- .neuralog.json|  <- runtime config|
|     |   +- ns/pod.log    |                   |
|     |   +- ns/pod.log.1  |  <- rotated files |
|     +--------------------+                   |
|                                              |
|  [CronJob: janitor]  (nightly retention)     |
+----------------------------------------------+
```

---

## Contributing

```bash
make test       # Go unit tests with race detector
make lint       # go vet + eslint
make helm-lint  # Helm strict lint
```

Open an issue before submitting significant changes. MIT License — see [`LICENSE`](LICENSE).
