# ── Stage 1: build ──
FROM node:20-alpine AS builder

WORKDIR /app

COPY ui/package.json ui/package-lock.json* ./
RUN npm ci --frozen-lockfile

COPY ui/ .
RUN npm run build

# ── Stage 2: nginx ──
FROM nginx:1.27-alpine

RUN addgroup -g 1001 -S nlui && adduser -u 1001 -S nlui -G nlui

COPY --from=builder /app/dist /usr/share/nginx/html
COPY docker/nginx.conf /etc/nginx/nginx.conf

RUN chown -R nlui:nlui /usr/share/nginx/html && \
    chown -R nlui:nlui /var/cache/nginx && \
    chown -R nlui:nlui /var/log/nginx && \
    touch /run/nginx.pid && chown nlui:nlui /run/nginx.pid

USER 1001

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
