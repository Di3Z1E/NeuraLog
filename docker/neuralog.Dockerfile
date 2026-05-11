# ── Stage 1: build UI ────────────────────────────────────────────────────────
FROM node:20-alpine AS ui-builder

WORKDIR /app

COPY ui/package.json ui/package-lock.json* ./
RUN npm ci --frozen-lockfile

COPY ui/ .
RUN npm run build

# ── Stage 2: build Go binary (UI dist embedded) ───────────────────────────────
FROM golang:1.25-alpine AS go-builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY collector/go.mod collector/go.sum ./
RUN go mod download

COPY collector/ .

# Embed the compiled UI into the Go binary at build time.
COPY --from=ui-builder /app/dist ./internal/ui/dist/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
    -ldflags="-s -w" \
    -o /bin/neuralog \
    ./cmd/neuralog

# ── Stage 3: minimal runtime ──────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-builder /bin/neuralog /neuralog

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/neuralog"]
CMD ["serve"]
