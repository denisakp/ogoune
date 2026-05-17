FROM node:24-alpine AS frontend-builder

WORKDIR /build/web

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV VITE_API_BASE_URL=/api
ENV CI=true
ENV PNPM_CONFIG_VERIFY_DEPS_BEFORE_RUN=false
ENV PNPM_CONFIG_CONFIRM_MODULES_PURGE=false

RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN printf 'verify-deps-before-run=false\nconfirm-modules-purge=false\n' > .npmrc

RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile --config.dangerouslyAllowAllBuilds=true
COPY web .
RUN printf 'verify-deps-before-run=false\nconfirm-modules-purge=false\n' > .npmrc

RUN pnpm build

# Stage 2: Build Go Backend
FROM golang:1.25-alpine3.22 AS go-builder

LABEL org.opencontainers.image.source="https://github.com/denisakp/ogoune"
LABEL org.opencontainers.image.description="A monitoring solution offering uptime monitoring, performance tracking, and alerting features."
LABEL org.opencontainers.image.title="Ogoune"

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY docs/ docs/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o ogoune ./cmd/api/main.go

# Stage 3: Final Image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S ogoune && \
    adduser -u 1001 -S ogoune -G ogoune

WORKDIR /app

RUN mkdir -p /data

# Copy the Go binary from go-builder stage
COPY --from=go-builder /build/ogoune .

# Copy the built Vue.js frontend static files from frontend-builder stage
COPY --from=frontend-builder /build/web/dist ./static

RUN chown -R ogoune:ogoune /app /data
USER ogoune

ENV PORT=9500
EXPOSE 9500

CMD ["./ogoune"]
