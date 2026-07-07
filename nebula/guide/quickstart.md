# Quickstart

Run Ogoune in Community mode (zero external dependencies).

## Prerequisites

- Go 1.24+
- A secret key: `openssl rand -hex 32`

## Run

```bash
cp .env.example .env          # set APP_SECRET_KEY
DB_DRIVER=sqlite \
  SQLITE_PATH=./ogoune.db \
  SCHEDULER_MODE=timingwheel \
  go run ./cmd/api
```

::: warning
`APP_SECRET_KEY` is mandatory — the app refuses to start without it.
:::

## Full stack (production-like)

```bash
docker compose up -d          # Postgres + Redis
go run ./cmd/api
```

## Frontend dev

```bash
cd web && pnpm install && pnpm dev   # http://localhost:5173
```

Next: [Core concepts](/guide/concepts).
