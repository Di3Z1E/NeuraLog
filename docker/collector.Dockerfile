# ── Stage 1: build ──
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY collector/go.mod collector/go.sum ./
RUN go mod download

COPY collector/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
    -ldflags="-s -w" \
    -o /bin/neuralog \
    ./cmd/neuralog

# ── Stage 2: runtime ──
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/neuralog /neuralog

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/neuralog"]
CMD ["serve"]
