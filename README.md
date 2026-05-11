<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,6,11,6,6,11,30,11,6&height=200&section=header&text=NeuraLog&fontSize=80&fontColor=818cf8&fontAlignY=45&desc=Kubernetes-Native%20Real-Time%20Log%20Aggregation&descSize=22&descColor=a5b4fc&descAlignY=70&animation=fadeIn" width="100%" />

<img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&size=22&duration=3000&pause=1000&color=818cf8&center=true&vCenter=true&width=640&lines=Real-Time+K8s+Log+Streaming+%E2%9A%A1;Configure+Everything+from+the+UI+%E2%9A%99%EF%B8%8F;Sensitive+Data+Redaction+%F0%9F%94%92;Log+Rotation+%C2%B7+Storage+Quota+%F0%9F%97%84%EF%B8%8F;Zero+Sidecars+%C2%B7+Single+Helm+Install+%F0%9F%94%8D" alt="Typing SVG" />

<br/>

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=1a1a1a)](https://go.dev)
[![Helm](https://img.shields.io/badge/Helm-3.x-0F1689?style=for-the-badge&logo=helm&logoColor=white&labelColor=1a1a1a)](helm/neuralog)
[![License](https://img.shields.io/github/license/Di3Z1E/neuralog?style=for-the-badge&color=818cf8&labelColor=1a1a1a)](LICENSE)

<br/>

<a href="#-quick-start"><b>Quick Start</b></a> • <a href="#%EF%B8%8F-settings--runtime-config"><b>Settings</b></a> • <a href="#-api"><b>API Reference</b></a> • <a href="#-redaction"><b>Redaction</b></a> • <a href="#%EF%B8%8F-helm-deployment"><b>Helm Deployment</b></a>

</div>

---

## ⚡ Overview

**NeuraLog** is a lightweight Kubernetes-native log aggregation platform. It discovers running pods via Kubernetes Informers, streams their logs to NFS storage with automatic sensitive-data redaction, and serves a dark-theme web UI with real-time WebSocket streaming, historical search, and a **built-in settings panel** to configure every operational parameter at runtime — no YAML editing, no pod restarts.

> [!WARNING]
> **v0.1.0 - actively stabilising.** The collector, store, and WebSocket APIs are functional but not yet covered by a stability guarantee. Pin image tags in production and review your NFS configuration before deploying.

### 🔥 Key Features

- ⚙️ **UI-Driven Runtime Config** — change storage quota, log rotation, retention, namespace exclusions, and redaction rules from the web UI; takes effect immediately without restarting
- ⚡ **Informer-Based Discovery** — pod events delivered in milliseconds via Kubernetes watch API, no polling
- 🔒 **Sensitive-Data Redaction** — JWT, Bearer tokens, AWS keys, passwords, DB connection strings, credit card numbers stripped before hitting disk or the wire; add custom regex patterns from the UI
- 🌐 **Live WebSocket Streaming** — real-time log tail with exponential-backoff reconnect and 10k-line virtual scroll
- 🗄️ **Log Rotation & Quota** — per-pod file rotation by size, configurable history depth, hard storage cap with oldest-first eviction
- 📦 **Zero Sidecars** — single Deployment, no DaemonSet, no admission webhooks, no mutations
- 🧹 **Automated Retention** — `janitor` subcommand runs as a CronJob; TTL is configurable from the UI
- 🛡️ **Hardened by Default** — `distroless/static:nonroot`, `readOnlyRootFilesystem`, dropped capabilities, NetworkPolicy

---

## 💎 Why NeuraLog?

- 🚫 **No heavy stack** — no Fluentd, no Fluentbit, no Loki, no Elasticsearch
- 🔍 **Audit-friendly** — redaction runs before any write; `[REDACTED:TYPE]` tokens are visible in the UI so you know exactly what was masked
- 🧩 **Single Helm install** — RBAC, PV/PVC, Ingress, HPA, NetworkPolicy all in one chart
- ⚙️ **Zero-restart reconfiguration** — config is persisted to `.neuralog.json` on the NFS volume and hot-reloaded; env vars remain supported as bootstrap values for the first boot
- 📊 **Scales horizontally** — shared NFS storage means multiple replicas read the same files; WebSocket clients connect to their pod replica independently

---

## 🛠️ Tech Stack

<div align="left">

**Backend** &nbsp;
![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![client-go](https://img.shields.io/badge/client--go-326CE5?style=flat-square&logo=kubernetes&logoColor=white)
![gorilla/websocket](https://img.shields.io/badge/gorilla%2Fwebsocket-333?style=flat-square)

**Frontend** &nbsp;
![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=flat-square&logo=typescript&logoColor=white)
![Vite](https://img.shields.io/badge/Vite-646CFF?style=flat-square&logo=vite&logoColor=white)

**Infrastructure** &nbsp;
![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=flat-square&logo=kubernetes&logoColor=white)
![Helm](https://img.shields.io/badge/Helm-0F1689?style=flat-square&logo=helm&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white)
![nginx](https://img.shields.io/badge/nginx-009639?style=flat-square&logo=nginx&logoColor=white)

</div>

---

## 🚀 Quick Start

### Prerequisites

- Docker + `docker compose`
- A `~/.kube/config` pointing at a live cluster (for local dev)
- Go 1.22+ (for building from source)
- Node 20+ (for UI development)

### Local Development

```bash
git clone https://github.com/Di3Z1E/neuralog
cd neuralog

# First build — generate go.sum
cd collector && go mod tidy && cd ..

# Start full stack with hot-reload
make dev
```

| Service | URL |
|---------|-----|
| UI (Vite dev server) | http://localhost:3000 |
| Collector API | http://localhost:8080 |

<details>
<summary><b>Run collector only (no Docker)</b></summary>

```bash
cd collector
go run ./cmd/neuralog serve
```

Environment variables act as **bootstrap defaults** for the first run. Once `.neuralog.json` exists on the storage volume, the UI-saved config takes precedence.

| Variable | Default | Description |
|----------|---------|-------------|
| `NEURALOG_LOG_BASE_PATH` | `/mnt/logs` | Base path for log storage |
| `NEURALOG_LISTEN_ADDR` | `:8080` | HTTP server address |
| `NEURALOG_EXCLUDE_NAMESPACES` | `log-system,kube-system` | Comma-separated namespaces to skip on first boot |
| `NEURALOG_REDACT_ENABLED` | `true` | Enable/disable redaction on first boot |
| `NEURALOG_RETENTION_DAYS` | `7` | Retention TTL on first boot |
| `KUBECONFIG` | _(in-cluster)_ | Path to kubeconfig for out-of-cluster runs |

</details>

<details>
<summary><b>Run UI only (Vite dev server)</b></summary>

```bash
cd ui
npm install
npm run dev   # → http://localhost:3000
```

The Vite dev proxy forwards `/api` and `/ws` to `localhost:8080` automatically.
</details>

---

## ⚙️ Settings & Runtime Config

Click the **gear icon** in the top-right of the UI to open the Settings panel. All changes are saved to `.neuralog.json` on the NFS volume and applied immediately — no pod restart required.

### Storage tab

| Setting | Description |
|---------|-------------|
| **Storage quota (GiB)** | Hard cap on total log disk usage. When exceeded the quota watcher evicts the oldest files. Set to `0` for unlimited. |
| **Rotation size (MB)** | Rotate a pod's `.log` file when it reaches this size. Rotated files are kept as `pod.log.1`, `pod.log.2`, … Set to `0` to disable. |
| **Rotated files to keep** | How many rotated files to retain per pod before the oldest is deleted. |
| **Retention (days)** | The nightly janitor CronJob deletes `.log` and `.log.N` files older than this many days. |

### Collection tab

Add or remove **excluded namespaces**. Newly excluded namespaces have their active log streams stopped immediately; newly included namespaces are picked up on the next informer event.

### Redaction tab

Toggle the **master redaction switch** on or off, and manage **custom regex patterns** that are applied on top of the built-in rules. Pattern changes take effect on the next incoming log line — no restart needed.

---

## ⎈ Helm Deployment

```bash
# 1. Build and push images
TAG=v0.1.0 make push

# 2. Install
helm upgrade --install neuralog helm/neuralog \
  --namespace log-system \
  --create-namespace \
  --set image.collector.tag=v0.1.0 \
  --set image.ui.tag=v0.1.0 \
  --wait
```

<details>
<summary><b>Key values to override</b></summary>

```yaml
# values.override.yaml
storage:
  size: 100Gi

ingress:
  enabled: true
  host: logs.yourdomain.com
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod

collector:
  redactEnabled: "true"
  excludeNamespaces: "log-system,kube-system,monitoring"

retention:
  days: 14
  schedule: "0 2 * * *"
```

```bash
helm upgrade --install neuralog helm/neuralog -f values.override.yaml \
  --namespace log-system --create-namespace
```

Full reference: [`helm/neuralog/values.yaml`](helm/neuralog/values.yaml)

> [!TIP]
> Env vars are only used to seed the config on the very first boot. After that, use the Settings UI — changes persist across pod restarts via `.neuralog.json` on the NFS volume.

</details>

---

## 📡 API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/pods` | List all tracked pods (live + historical) |
| `GET` | `/api/v1/logs/{namespace}/{pod}` | Fetch historical logs (`?lines=N&search=S&level=L&from=T&to=T`) |
| `GET` | `/api/v1/download/{namespace}/{pod}` | Download raw `.log` file |
| `GET` | `/api/v1/config` | Get current runtime config (includes `storageUsedGB`) |
| `PUT` | `/api/v1/config` | Update runtime config; hot-reloads redactor and collector exclusions |
| `WS`  | `/ws?namespace=N&pod=P` | Real-time log stream (seeds 200 history lines, then streams live) |
| `GET` | `/healthz` | Health check |

> [!TIP]
> The WebSocket endpoint sends the last 200 lines immediately upon connection, then streams new lines in real-time. Automatic reconnection with exponential backoff (1s → 30s) is handled by the frontend.

### Config schema (`PUT /api/v1/config`)

```json
{
  "storageQuotaGB":    10,
  "rotationMaxMB":     100,
  "rotationKeepFiles": 5,
  "retentionDays":     7,
  "excludeNamespaces": ["log-system", "kube-system"],
  "redactEnabled":     true,
  "customPatterns": [
    { "id": "uuid", "pattern": "my-secret-[a-z0-9]+", "replace": "[REDACTED:CUSTOM]" }
  ]
}
```

---

## 🔒 Redaction

All patterns run in the collector pipeline **before any write to disk or broadcast over WebSocket**. Redacted tokens appear visually distinct (`[REDACTED:TYPE]` in orange italic) in the UI.

| Pattern | Token |
|---------|-------|
| JWT tokens | `[REDACTED:JWT]` |
| Bearer tokens | `[REDACTED:BEARER_TOKEN]` |
| AWS key IDs (`AKIA`/`ASIA`/`AROA`) | `[REDACTED:AWS_KEY_ID]` |
| AWS secret keys | `[REDACTED:AWS_SECRET]` |
| Generic API keys | `[REDACTED:API_KEY]` |
| Passwords in log lines | `[REDACTED:PASSWORD]` |
| Generic secrets / tokens | `[REDACTED:SECRET]` |
| Database URLs with credentials | `[REDACTED:DB_URL]` |
| Basic auth in URLs | `[REDACTED:CREDENTIALS]` |
| Private key PEM blocks | `[REDACTED:PRIVATE_KEY]` |
| Credit card numbers | `[REDACTED:CARD]` |

Custom patterns can be added at runtime via the **Settings → Redaction** tab or via `PUT /api/v1/config`.

---

## 📐 Architecture

```
┌──────────────────────────────────────────────┐
│              Kubernetes Cluster              │
│                                              │
│  [Pod A]  [Pod B]  [Pod C]  ...              │
│      │        │        │                     │
│      └────────┴────────┘                     │
│               │  Informer watch              │
│               ▼                              │
│     ┌────────────────────┐                   │
│     │   NeuraLog Pod     │                   │
│     │                    │                   │
│     │  [collector] ◄─────┼── client-go       │
│     │       │            │                   │
│     │  redact + write    │                   │
│     │       │            │                   │
│     │  [NFS /mnt/logs]   │                   │
│     │   ├─ .neuralog.json│  ← runtime config │
│     │   ├─ ns/pod.log    │                   │
│     │   └─ ns/pod.log.1  │  ← rotated files  │
│     │       │            │                   │
│     │  REST + WS + Config│                   │
│     │       │            │                   │
│     │  [nginx + ui] ─────┼── Browser         │
│     └────────────────────┘                   │
│                                              │
│  [CronJob: janitor]  (nightly retention)     │
└──────────────────────────────────────────────┘
```

Two containers share the pod network — nginx proxies `/api/` and `/ws` to `localhost:8080` with WebSocket upgrade headers. No extra Service hop.

---

## 🔬 Comparison

| | **NeuraLog** | Bash/kubectl loop |
|---|:---:|:---:|
| Pod discovery latency | ~ms (Informer) | up to 30 s (poll) |
| Sensitive-data redaction | ✅ built-in | ❌ |
| Web UI | ✅ dark theme, virtual scroll | ❌ |
| Live WebSocket streaming | ✅ | ❌ |
| Reconnect on stream drop | ✅ automatic | ❌ manual restart |
| Runtime config (no restart) | ✅ UI + API | ❌ |
| Log rotation + storage quota | ✅ configurable | ❌ |
| Distroless / non-root | ✅ | ❌ runs as root |
| Helm chart | ✅ | ❌ raw YAML |

---

## 🤝 Contributing

Issues and pull requests are welcome. For significant changes, open an issue first to discuss the approach.

```bash
make test       # Go tests with race detector
make lint       # go vet + eslint
make helm-lint  # helm lint with required values
```

Distributed under the **MIT License** - see [`LICENSE`](LICENSE).

<div align="center">

---

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=6,6,11,6,6,11,30,11,6&height=100&section=footer" width="100%" />

*NeuraLog © 2026 Di3Z1E*

</div>
