FROM node:lts-alpine  AS frontend-builder

WORKDIR /build/frontend

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV NODE_ENV=production
ENV VITE_API_BASE_URL=/api

RUN corepack enable

COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile

COPY frontend .

RUN pnpm run build

# Stage 2: Build Go Backend
FROM golang:1.25.3-alpine3.22 AS go-builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download && go mod verify

COPY backend .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o pulseguard ./cmd/api/main.go

# Stage 3: Final Image
FROM alpine:latest

ENV DATABASE_URL= \
    REDIS_URL= \
    SMTP_HOST= \
    SMTP_PORT= \
    SMTP_USER= \
    SMTP_PASSWORD= \
    SMTP_SENDER= \
    DEFAULT_RECIPIENT_EMAIL= \
    STATIC_DIR=/app/static \
    APP_PORT=8080

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S pulsguard && \
    adduser -u 1001 -S pulsguard -G pulsguard

WORKDIR /app

# Copy the Go binary from go-builder stage
COPY --from=go-builder /build/pulseguard .

# Copy the built Vue.js frontend static files from frontend-builder stage
COPY --from=frontend-builder /build/frontend/dist ./static

RUN chown -R pulsguard:pulsguard /app
USER pulsguard

EXPOSE ${PORT}

CMD ["./pulseguard"]