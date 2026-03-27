# PulseGuard — Quick Start Guide

Everything you need to go from zero to a running PulseGuard instance.

---

## Table of Contents

- [Choose your setup](#choose-your-setup)
- [Option 1 — Community Edition (recommended)](#option-1--community-edition-recommended)
- [Option 2 — Full Stack (PostgreSQL + Redis)](#option-2--full-stack-postgresql--redis)
- [Option 3 — Build from source](#option-3--build-from-source)
- [First login](#first-login)
- [Step 1 — Add your first HTTP monitor](#step-1--add-your-first-http-monitor)
- [Step 2 — Configure notifications](#step-2--configure-notifications)
- [Environment variables reference](#environment-variables-reference)
- [Troubleshooting](#troubleshooting)

---

## Choose your setup

| | Community Edition | Full Stack | Build from source |
|---|---|---|---|
| External dependencies | None | PostgreSQL + Redis | None |
| Best for | Homelabs, solo devs, quick eval | Teams, production | Contributors, customization |
| Setup time | ~1 minute | ~5 minutes | ~10 minutes |
| Persistent data | SQLite file (volume mount) | PostgreSQL | SQLite or PostgreSQL |
| ARM / Raspberry Pi | ✅ | ✅ | ✅ |

---

## Option 1 — Community Edition (recommended)

Zero external dependencies. One container. Data stored in an embedded SQLite file.

### Prerequisites

- Docker 24+ installed and running

### Run

```bash
docker run -d \
  --name pulseguard \
  --restart unless-stopped \
  -p 8080:8080 \
  -v pulseguard_data:/data \
  -e JWT_SECRET=change-me-before-production \
  pulseguard/community:latest
```

Open **http://localhost:8080**

That's it. Your data is persisted in the `pulseguard_data` Docker volume.

### Or with Docker Compose

Create a `compose.yml`:

```yaml
services:
  pulseguard:
    image: pulseguard/community:latest
    container_name: pulseguard
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - pulseguard_data:/data
    environment:
      DB_DRIVER: sqlite
      SQLITE_PATH: /data/pulseguard.db
      SCHEDULER_DRIVER: timingwheel
      JWT_SECRET: change-me-before-production
      ADMIN_EMAIL: admin@pulseguard.test

volumes:
  pulseguard_data:
```

```bash
docker compose up -d
```

### Verify it's running

```bash
curl http://localhost:8080/health
# → {"status":"ok"}
```

### Verify status entrypoint routing

```bash
curl -i http://localhost:8080/status
curl -i http://localhost:8080/status/test-resource
curl -i http://localhost:8080/api/non-existent-route
```

Expected:
- `/status` and `/status/*` are served by the public status entry when `status.html` is present.
- unmatched `/api/*` routes return `404` and never static HTML.

---

## Option 2 — Full Stack (PostgreSQL + Redis)

For teams or production deployments that need a more robust backend.

### Prerequisites

- Docker 24+ and Docker Compose

### Setup

```bash
git clone https://github.com/denisakp/pulseguard.git
cd pulseguard
cp .env.example .env
```

Open `.env` and set at minimum:

```env
DATABASE_URL=postgres://pulseguard:pulseguard@postgres:5432/pulseguard
REDIS_URL=redis://redis:6379
JWT_SECRET=a-long-random-secret-string
ADMIN_EMAIL=admin@yourdomain.com
```

Generate a strong JWT secret:
```bash
openssl rand -hex 32
```

Start everything:

```bash
docker compose up -d
```

This starts: the PulseGuard app, PostgreSQL, Redis, and a reverse proxy.

### Verify services are healthy

```bash
docker compose ps
# All services should show "healthy" or "running"

curl http://localhost:8080/health
# → {"status":"ok"}
```

### Data persistence

PostgreSQL data is stored in the `postgres_data` Docker volume. Back it up regularly:

```bash
docker exec pulseguard-postgres pg_dump -U pulseguard pulseguard > backup.sql
```

---

## Option 3 — Build from source

For contributors or users who want to run the latest code.

### Prerequisites

| Tool | Version |
|---|---|
| Go | 1.24+ |
| Node.js | 22+ |
| Git | Latest |

### Clone and set up

```bash
git clone https://github.com/denisakp/pulseguard.git
cd pulseguard
```

### Build the frontend

```bash
cd web
pnpm install
pnpm build
# → builds to web/dist/
cd ..
```

Build output now contains both entry documents:
- `web/dist/index.html` for dashboard routes
- `web/dist/status.html` for public status routes

### Configure the backend

```bash
cp .env.example .env
```

Edit `.env` — minimum required for a local SQLite run:

```env
DB_DRIVER=sqlite
SQLITE_PATH=./pulseguard.db
SCHEDULER_MODE=timingwheel
JWT_SECRET=dev-secret-change-in-production
AUTH_EMAIL=admin@pulseguard.test
PORT=8080
STATIC_DIR=web/dist
```

### Run the backend

```bash
go run ./cmd/api
```

Open **http://localhost:8080**

The backend serves the frontend from `web/dist/` automatically.

### Run frontend in dev mode (hot reload)

In a second terminal:

```bash
cd web
pnpm dev
# → http://localhost:5173 (proxies API to :8080)
```

### Run tests

```bash
make test
```

### Build from the repo root

```bash
make build
```

This produces:
- `dist/pulseguard`
- `web/dist/index.html`
- `web/dist/status.html`

---

## First login

Regardless of which option you chose, the default credentials are:

| | |
|---|---|
| **URL** | http://localhost:8080 |
| **Email** | `admin@pulseguard.test` |
| **Password** | `puls3gu@rd` |

**Change your password immediately** — Settings → Account → Change password.

Optionally, enable 2FA — Settings → Account → Two-Factor Authentication.

---

## Step 1 — Add your first HTTP monitor

This walks you through adding a monitor for a website or API endpoint.

### 1.1 Open the monitors page

Click **Monitors** in the left sidebar, then **+ Add Monitor**.

### 1.2 Fill in the basic details

| Field | Example | Notes |
|---|---|---|
| **Name** | `My Website` | Displayed on the dashboard |
| **Type** | `HTTP` | Use TCP for non-HTTP services |
| **Target** | `https://example.com` | Full URL including protocol |
| **Interval** | `300` | Seconds between checks (default: 5 min) |
| **Timeout** | `10` | Seconds before a check is considered failed |

### 1.3 Configure the confirmation window

The confirmation window prevents false positives. PulseGuard will only create an incident after N consecutive failures.

| Field | Recommended | Notes |
|---|---|---|
| **Confirmation checks** | `2` | Failures before alerting |
| **Confirmation interval** | `30` | Seconds between confirmation checks |

> With these defaults: if your site goes down, PulseGuard waits 30 seconds and checks again. Only if it's still down does it create an incident and send an alert. A 2-second network blip will never wake you up at 3am.

Set **Confirmation checks to 1** if you want immediate alerts with no confirmation delay.

### 1.4 Save and verify

Click **Save**. The monitor appears on your dashboard with status **Pending** — it will run its first check within the configured interval.

After the first check, the status updates to:
- 🟢 **Up** — resource is reachable
- 🔴 **Down** — resource is unreachable (incident created after confirmation)
- 🟡 **Flapping** — resource is switching between up and down repeatedly

### 1.5 Add a TCP monitor (optional)

For services that aren't HTTP (databases, SMTP servers, game servers):

| Field | Example |
|---|---|
| **Type** | `TCP` |
| **Target** | `db.example.com:5432` |

Format: `hostname:port`. No protocol prefix.

### 1.6 Add a DNS monitor (optional)

To verify DNS resolution for a domain:

| Field | Example |
|---|---|
| **Type** | `DNS` |
| **Target** | `example.com` |

---

## Step 2 — Configure notifications

PulseGuard supports two notification channels out of the box: **SMTP (email)** and **Webhook** (Slack, Google Chat, Teams, Discord, or any HTTP endpoint).

### 2.1 Navigate to notification settings

Settings → Notifications → **+ Add Channel**

---

### SMTP — Email notifications

#### What you need

- An SMTP server (Gmail, Mailgun, Postmark, your own server, etc.)
- SMTP host, port, username, and password

#### Gmail setup

If you use Gmail, you need an **App Password** — Gmail blocks direct password auth for apps.

1. Go to [myaccount.google.com/security](https://myaccount.google.com/security)
2. Enable 2-Step Verification if not already enabled
3. Search for "App passwords" → Create → select "Mail" → copy the 16-char password

Use these settings:

| Field | Value |
|---|---|
| **Name** | `Gmail` (or anything) |
| **SMTP Host** | `smtp.gmail.com` |
| **Port** | `587` |
| **Username** | your full Gmail address |
| **Password** | the 16-char App Password |
| **From** | your Gmail address |
| **To** | where alerts should be sent |
| **TLS** | STARTTLS |
| **Enabled by default** | ✅ Toggle on |

#### Other providers

| Provider | Host | Port | TLS |
|---|---|---|---|
| Mailgun | `smtp.mailgun.org` | `587` | STARTTLS |
| Postmark | `smtp.postmarkapp.com` | `587` | STARTTLS |
| SendGrid | `smtp.sendgrid.net` | `587` | STARTTLS |
| Custom / self-hosted | your host | `25` or `587` | depends |

#### Test before saving

Click **Test** before saving. PulseGuard sends a test email immediately. If it arrives, the config is correct.

---

### Webhook — Slack, Google Chat, Teams, Discord

Webhooks work by sending an HTTP POST to a URL when an incident occurs. Every service listed below provides a webhook URL you paste into PulseGuard.

#### Slack

1. Go to [api.slack.com/apps](https://api.slack.com/apps) → **Create New App** → **From scratch**
2. Name it `PulseGuard`, choose your workspace
3. Click **Incoming Webhooks** → toggle **On**
4. Click **Add New Webhook to Workspace** → pick a channel → **Allow**
5. Copy the webhook URL — looks like: `https://hooks.slack.com/services/T.../B.../...`

In PulseGuard:

| Field | Value |
|---|---|
| **Name** | `Slack — #alerts` |
| **Type** | `Webhook` |
| **URL** | the URL copied from Slack |
| **Enabled by default** | ✅ Toggle on |

#### Google Chat

1. Open the Google Chat space where you want alerts
2. Click the space name → **Apps & integrations** → **Add webhooks**
3. Name it `PulseGuard` → **Save** → copy the URL

Paste the URL in PulseGuard exactly as you did for Slack.

#### Microsoft Teams

1. In Teams, right-click a channel → **Manage channel** → **Connectors**
2. Search for **Incoming Webhook** → **Add** → **Add** again
3. Name it `PulseGuard` → upload an icon (optional) → **Create**
4. Copy the webhook URL

#### Discord

1. In Discord, open channel settings → **Integrations** → **Webhooks**
2. Click **New Webhook** → name it → copy the URL
3. **Important:** append `/slack` to the URL — Discord's Slack-compatible webhook endpoint

```
https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN/slack
```

Paste this modified URL in PulseGuard.

#### Custom HTTP endpoint

Any endpoint that accepts a POST with a JSON body works. PulseGuard sends:

```json
{
  "event": "down",
  "resource_name": "Production API",
  "target": "https://api.example.com",
  "cause": "DNS Resolution Failed — host could not be found",
  "started_at": "2025-09-01T10:00:00Z"
}
```

---

### 2.2 Enable by default

When creating a notification channel, toggle **Enabled by default** to ON. This means the channel will receive alerts for **all monitors** — even ones that don't have a channel explicitly assigned.

If you leave this off, you need to manually assign the channel to each monitor individually.

### 2.3 Assign a channel to a specific monitor (optional)

To send alerts from a specific monitor to a specific channel only:

1. Open the monitor → **Edit**
2. Under **Notification Channels** → select the channel
3. Save

This overrides the default channel for that monitor.

### 2.4 Test the channel

Always click **Test** after creating a channel. It sends a test notification immediately. Confirm you received it before assuming it's working.

---

## Environment variables reference

Full list of all available configuration options.

```env
# ── Database ────────────────────────────────────────────────────────────────
DB_DRIVER=sqlite              # sqlite | postgres
SQLITE_PATH=/data/pulseguard.db  # Path to SQLite file (Community Edition)
DATABASE_URL=postgres://user:pass@host:5432/dbname  # Required if postgres
DB_LOG_LEVEL=error            # silent | error | warn | info

# ── Scheduler ────────────────────────────────────────────────────────────────
SCHEDULER_DRIVER=timingwheel  # timingwheel | asynq
REDIS_URL=redis://localhost:6379  # Required if SCHEDULER_DRIVER=asynq
SCHEDULER_CONCURRENCY=10      # Max parallel checks
SCHEDULER_NOTIF_QUEUE_SIZE=100  # Notification queue buffer

# ── Alerting defaults ────────────────────────────────────────────────────────
CONFIRMATION_CHECKS=2         # Failures before alerting (per monitor override available)
CONFIRMATION_INTERVAL=30      # Seconds between confirmation checks
EXPIRY_ALERT_THRESHOLDS=30,14,7,1  # Days before SSL/domain expiry to alert

# ── Application ──────────────────────────────────────────────────────────────
APP_PORT=8080
APP_ENV=production            # development | production

# ── Security ─────────────────────────────────────────────────────────────────
JWT_SECRET=                   # Required — use `openssl rand -hex 32` to generate
ADMIN_EMAIL=admin@pulseguard.test  # Default admin account email
```

---

## Troubleshooting

### Container starts but I can't open the UI

```bash
# Check the container is running
docker ps | grep pulseguard

# Check logs for errors
docker logs pulseguard

# Verify the health endpoint
curl http://localhost:8080/health
```

Common causes:
- Port 8080 already in use → change `-p 8080:8080` to `-p 9090:8080`
- `JWT_SECRET` not set → add `-e JWT_SECRET=any-random-string`

---

### My monitor shows "Pending" and never updates

The first check runs after the configured interval. For a 5-minute interval, wait 5 minutes.

To trigger a check immediately: pause the monitor and resume it — this reschedules the check immediately.

---

### I'm not receiving notifications

**Step 1 — Check channel configuration:**
Settings → Notifications → click your channel → **Test**. If the test fails, the SMTP or webhook config is wrong.

**Step 2 — Check "Enabled by default":**
If the channel is not set as default and not assigned to the monitor, no notification is sent. Edit the channel and toggle **Enabled by default** on.

**Step 3 — Check PulseGuard logs:**
```bash
docker logs pulseguard | grep "NOTIFICATION"
```
Look for `[WARNING] No notification channels configured` — this means the monitor has no channel assigned and no default is set.

**Step 4 — Check the incident event steps:**
Open the incident detail page. The timeline shows whether an alert was sent. If `alert_sent` step is missing, no notification was dispatched.

---

### Gmail SMTP returns "authentication failed"

You must use an **App Password**, not your regular Gmail password. See [Step 2 — Gmail setup](#gmail-setup).

---

### SQLite database file not persisting between restarts

You must mount a volume. Without `-v pulseguard_data:/data`, data is lost when the container stops.

```bash
# Correct
docker run -v pulseguard_data:/data pulseguard/community:latest

# Wrong — data lost on restart
docker run pulseguard/community:latest
```

---

### Build from source fails — "go: command not found"

Go is not installed or not in your PATH. Download from [go.dev/dl](https://go.dev/dl/) and follow the installation instructions for your OS.

---

### `go test -race ./...` fails with "too many open files" on macOS

```bash
# Increase the file descriptor limit
ulimit -n 10000
go test -race ./...
```

---

## Next steps

- [Architecture documentation](./ARCHITECTURE.md) — how PulseGuard works under the hood
- [Contributing guidelines](./CONTRIBUTING.md) — how to contribute code or feedback
- [GitHub Discussions](https://github.com/denisakp/pulseguard/discussions) — ask questions
- [GitHub Issues](https://github.com/denisakp/pulseguard/issues) — report bugs

---

*Found something wrong or missing in this guide? [Open an issue](https://github.com/denisakp/pulseguard/issues/new?labels=documentation) — doc fixes are always welcome.*