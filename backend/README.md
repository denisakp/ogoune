# Pulseguard - Backend API & Worker

The Pulseguard backend is a pure JSON API with a decoupled background worker for executing health checks. It powers the monitoring engine, incident lifecycle, notifications, and real-time updates.

## Tech Stack

- Go 1.25+
- Chi (HTTP router)
- GORM (PostgreSQL)
- Redis + Asynq (job queue)
- nhooyr.io/websocket (real-time)

## Configuration

Create an environment file based on the example:

```bash
cp backend/.env.example backend/.env
```

Required variables:

```env
# Server
PORT=8080
APP_ENV=development

# Database
DATABASE_URL=postgres://postgres:password@localhost:5432/pulseguard?sslmode=disable

# Redis (jobs)
REDIS_URL=localhost:6379

# SMTP (optional system notifications)
SMTP_HOST=
SMTP_PORT=
SMTP_USER=
SMTP_PASSWORD=
SMTP_SENDER=
DEFAULT_RECIPIENT_EMAIL=
```

## Running the Backend

Using Makefile helpers at repo root:

```bash
make docker-up     # starts Postgres and Redis
make run           # runs API + Worker in a single process
```

Or directly with Go:

```bash
cd backend
go run ./cmd/api
```

The API listens on `http://localhost:${PORT}` (default 8080).

## Developer Notes

- API handlers must only enqueue work; checks run in the worker.
- PostgreSQL is the source of truth. Redis is transport for jobs.
- Prefer small, testable functions in `internal/domain`.

## Architecture & Deep Dives

See `backend/docs/ARCHITECTURE.md` for the definitive backend architecture: API+Worker, scheduler (ticker), WebSockets hub, incident rules, and notification system.

## Contribution

- Run tests: `make test`
- Format: `gofmt -l -s -w .`
- Submit PRs with focused changes.
