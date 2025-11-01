FROM golang:1.25.3-alpine3.22 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o pulseguard .
   
FROM alpine:latest

ENV GIN_MODE=release \
    DATABASE_URL= \
    REDIS_URL= \
    SMTP_HOST= \
    SMTP_PORT= \
    SMTP_USER= \
    SMTP_PASSWORD= \
    SMTP_SENDER= \
    DEFAULT_RECIPIENT_EMAIL= 

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S pulsguard && \
    adduser -u 1001 -S pulsguard -G pulsguard

WORKDIR /app

COPY --from=builder /build/pulseguard .

RUN chown -R pulsguard:pulsguard /app
USER pulsguard

EXPOSE 8080

CMD ["./pulseguard"]