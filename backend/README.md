# Pulseguard Backend

The backend is a unified Go service that monitors resources (HTTP/TCP), detects incidents, and sends notifications via SMTP and webhooks. It exposes a pure JSON API and runs a Redis/Asynq-based background worker.

## Quick Links

- **Setup & Configuration** – See "Running Locally" and "Configuration" sections below
- **API Reference** – See "REST API" section below
- **Troubleshooting** – See "Troubleshooting" section below
- **Technical Deep Dive** – See [Backend Architecture](./ARCHITECTURE.md)
- **Developer Overview** – See [Backend Overview](./BACKEND_OVERVIEW.md)

---

## Prerequisites

- Go 1.24+
- Redis 6+
- PostgreSQL 16+ for hosted deployments, or a writable filesystem for SQLite community deployments
- Docker Compose (recommended for local development)

---

## Configuration

### Environment Variables

#### Core Settings

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `PORT` | HTTP server port | No | `8080` |
| `DB_DRIVER` | Database runtime (`postgres` or `sqlite`) | No | `postgres` |
| `DATABASE_URL` | PostgreSQL connection string | When `DB_DRIVER=postgres` | N/A |
| `SQLITE_PATH` | SQLite database path | No | `pulseguard.db` |
| `DB_LOG_LEVEL` | GORM log verbosity (`silent`, `error`, `warn`, `info`) | No | `error` |
| `REDIS_URL` | Redis connection string | No | `localhost:6379` |
| `APP_ENV` | Environment (development or production) | No | `development` |
| `STATIC_DIR` | Path to frontend static files | No | `./static` |

#### Notification Channels (Email / Slack / Webhook)

Notifications are now configured entirely through notification channels stored in the database. There is **no default SMTP configuration** from environment variables. Add channels via the UI or API (`/notification-channels`) and test them with `/notification-channels/{id}/test` or `/notification-channels/test-config`.

#### Webhook Notifications (HTTP Callbacks)

If you still want a global webhook fallback, you can keep `WEBHOOK_URL`/`WEBHOOK_SIGNATURE`, but prefer channel-based configuration for consistency with SMTP and Slack.

### Example .env File

```bash
# Core
PORT=8080
DB_DRIVER=sqlite
SQLITE_PATH=./pulseguard.db
DB_LOG_LEVEL=error
DATABASE_URL=postgres://postgres:password@localhost:5432/pulseguard?sslmode=disable
REDIS_URL=localhost:6379
APP_ENV=development
STATIC_DIR=./static

# Optional webhook fallback (prefer channel-based configuration)
WEBHOOK_URL=https://webhook.site/unique-id
WEBHOOK_SIGNATURE=my-secret-key
```

---

## Running Locally

### 1. Choose a database mode

Community / embedded SQLite:

```bash
DB_DRIVER=sqlite SQLITE_PATH=./pulseguard.db go run ./cmd/api
```

Hosted / PostgreSQL:

### 2. Start PostgreSQL and Redis

Using Docker Compose (recommended):

```bash
docker compose up -d
```

Or manually with Docker:

```bash
# PostgreSQL
docker run --rm \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=pulseguard \
  -p 5432:5432 \
  postgres:17-alpine

# Redis (in another terminal)
docker run --rm -p 6379:6379 redis:7
```

### 3. Set Up Environment

```bash
cp .env.example .env
# Edit .env with your settings
```

### 4. Run the Backend

```bash
go run ./cmd/api
```

Expected output:
```
✓ Database connection established successfully
✓ All systems operational!
✓ API server listening on :8080
```

The API is now available at `http://localhost:8080/api`

### 5. Verify the Setup

```bash
# Health check
curl http://localhost:8080/health

# List resources (should be empty initially)
curl http://localhost:8080/api/resources
```

---

## REST API

### Health & Status

```
GET /health                          Service health check
GET /api/status                      System status and incidents
GET /api/status/{resourceId}         Individual resource status
```

### Resources (Monitors)

```
GET    /api/resources                List all resources
POST   /api/resources                Create new resource
GET    /api/resources/{id}           Get resource details
PATCH  /api/resources/{id}           Update resource
DELETE /api/resources/{id}           Delete resource
POST   /api/resources/{id}/pause     Pause monitoring
POST   /api/resources/{id}/resume    Resume monitoring
POST   /api/resources/{resourceId}/tags        Add tags
DELETE /api/resources/{resourceId}/tags/{tagId} Remove tag
GET    /api/resources/{resourceId}/uptime-stats Uptime stats
```

### Incidents

```
GET /api/incidents                   List incidents (supports ?unresolved=true)
GET /api/incidents/{id}              Get incident details
GET /api/incidents/{id}/event-steps  Get incident event timeline
```

### Monitoring Activities

```
GET /api/monitoring-activities       Activity log (supports ?resource_id=...)
```

### Statistics

```
GET /api/stats/summary?range=24h     Global stats (24h, 7d, 30d, 90d)
```

### Notifications

```
POST /api/notifications/test         Send test email (requires SMTP configured)
```

### Tags

```
GET    /api/tags                     List all tags
POST   /api/tags                     Create new tag
PATCH  /api/tags/{id}                Update tag
DELETE /api/tags/{id}                Delete tag
```

---

## Creating Your First Monitor

### HTTP Monitor

```bash
curl -X POST http://localhost:8080/api/resources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Website",
    "type": "http",
    "target": "https://example.com",
    "interval": 60,
    "timeout": 10
  }'
```

### TCP Monitor

```bash
curl -X POST http://localhost:8080/api/resources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Database Server",
    "type": "tcp",
    "target": "db.example.com:5432",
    "interval": 60,
    "timeout": 5
  }'
```

After 3 consecutive failures, an incident is created and notifications are sent (if SMTP or webhook is configured).

---

## Testing Notifications

Use notification channels:
- Configure a channel via `/notification-channels` (SMTP, Slack, Webhook, SMS).
- Test a saved channel: `POST /notification-channels/{id}/test`.
- Validate config before saving: `POST /notification-channels/test-config`.

For ad-hoc webhook testing, you can still point a channel to a [webhook.site](https://webhook.site) URL.

---

## Troubleshooting

### Database Connection Fails

**Error:** `failed to initialize database`

**Solutions:**
- Verify `DATABASE_URL` format: `postgres://user:password@host:port/dbname?sslmode=disable`
- Verify `DB_DRIVER` matches the intended runtime (`postgres` or `sqlite`)
- For SQLite mode, ensure the parent directory of `SQLITE_PATH` is writable
- Ensure PostgreSQL is running: `docker ps` or `pg_isready -h localhost`
- Check credentials by connecting directly: `psql postgresql://user:password@localhost:5432/pulseguard`
- Verify network connectivity to the database host

### SQLite Startup Fails

- Verify `SQLITE_PATH` points to a writable location.
- Remove any partially created test database and restart after fixing the path or permissions.
- Review startup logs for migration failure details. The service intentionally fails fast before serving requests.

## Scope Notes

- SQLite support is intended for fresh community deployments with no external database.
- Automatic PostgreSQL-to-SQLite data migration is out of scope for this feature.
- Redis remains required for the scheduler and worker even when SQLite is used.

### Redis Connection Fails

**Error:** `could not connect to redis`

**Solutions:**
- Verify `REDIS_URL` (format: `host:port`, e.g., `localhost:6379`)
- Ensure Redis is running: `docker ps` or `redis-cli PING`
- Check network connectivity to the Redis host
- Check firewall rules if connecting to a remote Redis

### Notification Channels Not Sending

- Ensure the channel config is valid (use `POST /notification-channels/test-config`).
- For SMTP channels, double-check host/port/auth inside the saved channel JSON.
- For webhook channels, confirm the endpoint returns 2xx and accepts your payload.
- Review logs for the specific notifier (SMTP/Slack/Webhook) when incidents fire.

### Webhooks Not Triggering

**Error in logs:** `[WEBHOOK ERROR] failed to send webhook notification`

**Solutions:**
1. Verify `WEBHOOK_URL` is set and accessible:
   ```bash
   curl -X POST $WEBHOOK_URL -d '{"test": true}'
   ```
2. Ensure your webhook endpoint returns a 2xx HTTP status code
3. Check firewall allows outbound HTTPS (port 443)
4. Test with [webhook.site](https://webhook.site):
   - Generate a unique URL
   - Set it as `WEBHOOK_URL`
   - Trigger an incident
   - Check webhook.site for the request
5. Verify `WEBHOOK_SIGNATURE` (if set) is correct on your receiving end
6. Check application logs for detailed error messages

### No Incidents Created

**Symptoms:** Resources fail but no incidents appear

**Solutions:**
- Resource must fail **3 consecutive checks** (not just once)
- Each check runs at the resource's configured interval
  - E.g., with 60-second interval, 3 failures = ~3 minutes until incident
- Verify checks are running:
  ```bash
  curl http://localhost:8080/api/monitoring-activities
  ```
- If no activities appear, resource might be paused or not active
- Check logs for any check execution errors

### API Returns 404

**Problem:** Endpoints return 404 Not Found

**Solutions:**
- Verify API URL includes `/api/`: `http://localhost:8080/api/resources` ✅ not `http://localhost:8080/resources` ❌
- Ensure backend is running: `curl http://localhost:8080/health`
- Check that the endpoint is correct (see REST API section above)

### Application Won't Start

**Error:** `Error binding to port` or `address already in use`

**Solutions:**
- Port 8080 is already in use
- Change `PORT` in `.env` to an available port (e.g., 8081)
- Or find and stop the process using port 8080:
  ```bash
  lsof -i :8080      # Show process using port 8080
  kill -9 <PID>      # Terminate the process
  ```

---

## Development

### Running Tests

```bash
go test ./...          # Run all tests
go test -v ./...       # Verbose output
go test -cover ./...   # Coverage report
go test -run TestName  # Run specific test
```

### Building for Production

```bash
go build -o pulseguard-api ./cmd/api
./pulseguard-api
```

### Docker Build

```bash
docker build -f Dockerfile -t pulseguard-api .
docker run \
  -e DATABASE_URL=postgres://... \
  -e REDIS_URL=redis.example.com:6379 \
  pulseguard-api
```

---

## Production Deployment

### Pre-Deployment Checklist

- [ ] Set `APP_ENV=production`
- [ ] Configure PostgreSQL with strong password and restricted network access
- [ ] Configure Redis with authentication and restricted network access
- [ ] Set up SMTP for production email service (or disable if not needed)
- [ ] Configure `WEBHOOK_URL` for incident notifications
- [ ] Enable database backups (daily or more frequently)
- [ ] Use HTTPS for all external communication (reverse proxy or ingress)
- [ ] Set up monitoring and alerting for the backend service itself
- [ ] Run only one scheduler instance (use external orchestration if needed)
- [ ] Consider rate limiting and request validation (reverse proxy or middleware)

### Horizontal Scaling

- **API Server:** Deploy multiple instances behind a load balancer (stateless)
- **Scheduler:** Run only one instance to avoid duplicate scheduling
- **Worker:** Deploy multiple instances for higher throughput (share Redis queue)
- **Database:** Use PostgreSQL connection pooling (configured in code)
- **Redis:** Deploy as cluster or sentinel for high availability

### Performance Tuning

- **Worker Concurrency:** Adjust in `internal/worker/processor.go` (default: 10)
- **Database Pool:** Adjust in `internal/repository/postgres/database/init.go`
- **Slow Query Logging:** Enable in GORM config to identify bottlenecks
- **Caching:** Consider reverse proxy caching for `/status` and `/stats` endpoints

---

## Directory Structure

```
backend/
├── cmd/api/
│   └── main.go                       Application entry point
├── internal/
│   ├── api/
│   │   ├── router.go                 HTTP routes and CORS
│   │   └── handler/                  HTTP handlers
│   ├── config/
│   │   └── config.go                 Environment configuration
│   ├── domain/
│   │   ├── models.go                 Domain entities
│   │   └── check.go                  Check strategies
│   ├── monitoring/
│   │   ├── incident_service.go       Incident lifecycle
│   │   ├── scheduler_service.go       Task scheduling
│   │   └── strategy/                 HTTP and TCP strategies
│   ├── repository/
│   │   ├── interfaces.go             Repository contracts
│   │   ├── postgres/                 Database implementations
│   │   └── fake/                     Test doubles
│   ├── service/
│   │   ├── resource_service.go       Resource management
│   │   ├── notification_service.go   Manual notifications
│   │   └── ...                       Other services
│   └── worker/
│       ├── processor.go              Asynq worker setup
│       └── handler_monitoring.go     Check task handler
├── pkg/notifier/
│   ├── smtp.go                       Email notifications
│   ├── webhook.go                    Webhook notifications
│   ├── notifier.go                   Notifier interface
│   └── factory.go                    Notifier creation
├── docs/
│   ├── ARCHITECTURE.md               Technical deep dive
│   └── BACKEND_OVERVIEW.md           Developer guide
├── .env.example                      Example configuration
├── docker-compose.yml                Local development setup
└── Dockerfile                        Production image
```

---

## Contributing

- Follow [Effective Go](https://golang.org/doc/effective_go) style guidelines
- Add tests for new features
- Update documentation in `docs/` for architectural changes
- Keep commits focused and well-messaged

For detailed architecture and design patterns, see [Backend Architecture](./docs/ARCHITECTURE.md).

---

## License

MIT – see [LICENSE](../LICENSE)