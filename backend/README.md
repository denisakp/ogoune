# Pulseguard — Backend (API, Scheduler, Worker)

The backend monitors resources (HTTP/TCP), stores results in PostgreSQL, detects incidents, and sends notifications through SMTP and chat integrations. It exposes a pure JSON API and runs a Redis/Asynq–based scheduler and worker in the same process.

Highlights
- Single binary with three components: HTTP API server, Asynq periodic scheduler, Asynq worker
- PostgreSQL (GORM) as source of truth, Redis only for job scheduling/transport
- Incident lifecycle with event steps and two-layer notifications (system SMTP + integrations)
- Pre-aggregated status page data (90-day model with “no_data” before creation) and global stats

Useful docs
- Unified overview: docs/BACKEND_OVERVIEW.md
- Architecture: docs/ARCHITECTURE.md

---

Backend structure

- cmd/api/main.go
  Entry point: wires config, DB, Asynq (client/inspector/scheduler), repositories, services, handlers, worker, and router
- internal/api/
  - router.go: routes, middleware (CORS), content-type
  - handler/*: JSON handlers (resources, incidents, status page, stats, integrations, activities, notifications)
- internal/config/
  - config.go: env loading (dotenv in dev), SMTP enablement flag, required checks
- internal/domain/
  - models.go: entities (Resource, Incident, IncidentEventStep, Integration, NotificationEvent, MonitoringActivity, Tags)
  - check.go: strategies interface and check executor
- internal/monitoring/
  - scheduler_service.go: Asynq periodic registration per resource (“monitor:<id>” with @every <interval>s)
  - incident_service.go: 3-fail rule, resolve-on-recovery, notification fan-out and event steps
  - strategy/http.go, strategy/tcp.go: concrete check strategies
- internal/repository/postgres/
  - database/*: DB initialization, connection pooling, auto-migrations
  - repositories for each entity; plus aggregate queries (uptime, incident stats)
- internal/service/
  - orchestration logic for API handlers (resources, status page 90-day calc, stats summary, etc.)
- internal/worker/
  - processor.go: Asynq server and mux
  - handler_monitoring.go: “monitoring:check” task handler (execute → persist activity → state transitions → incident/notify)
- pkg/notifier/
  - smtp.go (+ templates), slack.go, discord.go, googlechat.go, factory.go

---

Dependencies

Required
- Go 1.24+
- PostgreSQL 16+ (DATABASE_URL)
- Redis 6+ (REDIS_URL)

Optional (enables system SMTP notifications)
- SMTP_HOST
- SMTP_PORT
- SMTP_USER
- SMTP_PASSWORD
- SMTP_SENDER
- DEFAULT_RECIPIENT_EMAIL

---

Configuration

Environment variables (examples in parentheses)
- PORT (8080)
- DATABASE_URL (postgres://user:password@localhost:5432/pulseguard?sslmode=disable)
- REDIS_URL (localhost:6379)

SMTP (all required to enable default/system SMTP notifications)
- SMTP_HOST
- SMTP_PORT
- SMTP_USER
- SMTP_PASSWORD
- SMTP_SENDER
- DEFAULT_RECIPIENT_EMAIL

Notes
- A local .env file is loaded automatically in development if present
- DATABASE_URL is required; the app exits if missing
- SMTP is enabled only when all SMTP_* variables above are set and non-empty

---

Running locally

1) Ensure PostgreSQL and Redis are running and reachable via DATABASE_URL / REDIS_URL
- Example Redis: `docker run --rm -p 6379:6379 redis:7`
- Example Postgres:
  - `docker run --rm -e POSTGRES_PASSWORD=password -e POSTGRES_USER=postgres -e POSTGRES_DB=pulseguard -p 5432:5432 postgres:17-alpine`
  - DATABASE_URL: `postgres://postgres:password@localhost:5432/pulseguard?sslmode=disable`
Or you can simply rely on the existing `docker-compose.yml` file

2) Export environment variables (or create a local .env)
- PORT=8080
- APP_ENV=development
- DATABASE_URL=postgres://postgres:password@localhost:5432/pulseguard?sslmode=disable
- REDIS_URL=localhost:6379
- Optional SMTP_* variables as above

3) Start the backend
- Change directory: `cd backend`
- Run: `go run ./cmd/api`
- On startup, the app will:
  - Connect to PostgreSQL, configure pool, and auto-migrate models
  - Connect to Redis, start the Asynq scheduler and worker
  - Bootstrap scheduling for active resources
  - Serve the dashboard at http://localhost:${PORT} and the JSON API under http://localhost:${PORT}/api

4) Useful endpoints (JSON)
- Health: GET /api/health
- Status page: GET /api/status, GET /api/status/{resourceId}
- Stats summary: GET /api/stats/summary?range=2h|24h|7d|30d
- Resources: GET/POST /api/resources, GET/PATCH/DELETE /api/resources/{id}, pause/resume, tags, uptime-stats
- Activities: GET /api/monitoring-activities (?resource_id=...)
- Incidents: GET /api/incidents (?unresolved=...), GET /api/incidents/{id}, GET /api/incidents/{id}/event-steps
- Integrations: GET/POST /api/integrations, PATCH /api/integrations/{id}
- Notifications: POST /api/notifications/test

---

How it works (short)

- Scheduling
  - On boot, each active resource is registered as an Asynq periodic entry (“monitor:<id>”) with `@every <interval>s`
  - Resource updates trigger re-scheduling (unregister + register)
- Execution
  - “monitoring:check” tasks are consumed by the worker (concurrency 10; queue “monitoring”)
  - The handler loads the resource, executes HTTP/TCP check with timeouts, and persists a MonitoringActivity
- State transitions and incidents
  - On consecutive fails: increment failure_count; on third failure, create incident if none active
  - On recovery: resolve the most recent active incident
  - Event steps are recorded for detection, resolution, and notifications
- Notifications
  - Layer 1: system SMTP (optional; global admin recipient)
  - Layer 2: integrations (Slack/Discord/Google Chat) filtered by subscribed event types
- Status and stats
  - `/api/status`: pre-aggregated 90-day per-resource series including “no_data” before creation
  - `/api/stats/summary`: global uptime %, incident counts, affected monitors in time windows

---

Troubleshooting

- Database connection fails
  - Verify DATABASE_URL (host/port/db/user/password)
  - Ensure Postgres is running and reachable (psql or `pg_isready`)
- Redis connection fails
  - Verify REDIS_URL (host:port)
  - Ensure Redis is running (`redis-cli PING`)
- SMTP not sending
  - SMTP is enabled only if all SMTP_* vars are set
  - Check logs for SMTP dial/auth errors
- Duplicate scheduling
  - Current binary starts a scheduler in every instance; prefer a single scheduler instance in production until a toggle/leader strategy is added

---

Development notes

- JSON-only API with permissive CORS by default (restrict origins in production)
- GORM logs slow queries; tune DB pool as needed
- Add new check strategies by implementing `CheckStrategy` and registering in `CheckExecutor`
- Add new notifiers by implementing the `Notifier` interface and wiring into the factory
