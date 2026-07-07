# Production (Postgres + Redis)

For horizontal scale, run Ogoune against **PostgreSQL** with an **Asynq (Redis)** scheduler. The API becomes stateless and check execution moves to external worker processes.

## Stack

```bash
docker compose up -d          # Postgres + Redis
./ogoune                      # API
```

## Configuration

```bash
DB_DRIVER=postgres
SCHEDULER_MODE=asynq
# + Postgres DSN and Redis connection — see Configuration
```

See [Configuration](/self-host/configuration) for the full environment reference.
