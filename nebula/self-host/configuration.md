# Configuration

Ogoune is configured via environment variables (`.env` supported).

## Core

| Variable | Required | Description |
|---|---|---|
| `APP_SECRET_KEY` | ✅ | 32-byte hex. App refuses to start without it. Generate with `openssl rand -hex 32`. |
| `DB_DRIVER` | ✅ | `sqlite` or `postgres` |
| `SCHEDULER_MODE` | ✅ | `timingwheel` (in-process) or `asynq` (Redis) |

## SQLite (Community)

| Variable | Description |
|---|---|
| `SQLITE_PATH` | Path to the SQLite database file |

## Postgres + Redis (Production)

Provide the Postgres DSN and Redis connection details. See `.env.example` in the repository for the complete, up-to-date list.

::: tip
The authoritative reference is always `.env.example` in the repo — it tracks new options as they land.
:::
