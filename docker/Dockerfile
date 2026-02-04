# Frontend build stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package*.json ./frontend/
RUN cd /app/frontend && npm ci

COPY frontend/ /app/frontend/
RUN cd /app/frontend && npm run build

# Backend build stage
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

COPY backend/ /app/backend/
WORKDIR /app/backend
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage - combined image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs nginx supervisor

WORKDIR /app

# Copy backend binary and migrations
COPY --from=backend-builder /app/backend/server /app/server
COPY --from=backend-builder /app/backend/migrations /app/migrations

# Copy frontend assets
COPY --from=frontend-builder /app/frontend/dist /usr/share/nginx/html

# Create nginx config directly
RUN cat > /etc/nginx/conf.d/default.conf << 'EOF'
server {
    listen 8088;
    server_name localhost;

    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /v1 {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;
        proxy_cache off;
    }
}
EOF

# Copy supervisord config
RUN mkdir -p /var/log/supervisor
COPY docker/supervisord.conf /etc/supervisor.d/supervisord.conf

# Create data directory
RUN mkdir -p /data

EXPOSE 8088

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor.d/supervisord.conf"]
