# Ogoune — Quick Start Guide

Everything you need to go from zero to a running Ogoune instance.

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
  --name ogoune \
  --restart unless-stopped \
  -p 8080:8080 \
  -v ogoune_data:/data \
  -e JWT_SECRET=change-me-before-production \
  ogoune/community:latest
```

Open **http://localhost:8080**

That's it. Your data is persisted in the `ogoune_data` Docker volume.

### Or with Docker Compose

Create a `compose.yml`:

```yaml
services:
  ogoune:
    image: ogoune/community:latest
    container_name: ogoune
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ogoune_data:/data
    environment:
      DB_DRIVER: sqlite
      SQLITE_PATH: /data/ogoune.db
      SCHEDULER_DRIVER: timingwheel
      JWT_SECRET: change-me-before-production
      ADMIN_EMAIL: admin@ogoune.test

volumes:
  ogoune_data:
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
git clone https://github.com/denisakp/ogoune.git
cd ogoune
cp .env.example .env
```

Open `.env` and set at minimum:

```env
DATABASE_URL=postgres://ogoune:ogoune@postgres:5432/ogoune
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

This starts: the Ogoune app, PostgreSQL, Redis, and a reverse proxy.

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
docker exec ogoune-postgres pg_dump -U ogoune ogoune > backup.sql
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
git clone https://github.com/denisakp/ogoune.git
cd ogoune
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
SQLITE_PATH=./ogoune.db
SCHEDULER_MODE=timingwheel
JWT_SECRET=dev-secret-change-in-production
AUTH_EMAIL=admin@ogoune.test
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
- `dist/ogoune`
- `web/dist/index.html`
- `web/dist/status.html`

---

## First login

Regardless of which option you chose, the default credentials are:

| | |
|---|---|
| **URL** | http://localhost:8080 |
| **Email** | `admin@ogoune.test` |
| **Password** | `ogu3n3@rd` |

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

The confirmation window prevents false positives. Ogoune will only create an incident after N consecutive failures.

| Field | Recommended | Notes |
|---|---|---|
| **Confirmation checks** | `2` | Failures before alerting |
| **Confirmation interval** | `30` | Seconds between confirmation checks |

> With these defaults: if your site goes down, Ogoune waits 30 seconds and checks again. Only if it's still down does it create an incident and send an alert. A 2-second network blip will never wake you up at 3am.

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

### 1.7 Add an ICMP monitor (optional)

ICMP monitoring is opt-in because raw ICMP sockets depend on host/container capabilities that are not universally available by default.

Enable it in `.env` first:

```env
ENABLE_ICMP=true
```

Then restart the app and open **Monitors → + Add Monitor**.

| Field | Example | Notes |
|---|---|---|
| **Type** | `ICMP` | Ping-based reachability check |
| **Target** | `1.1.1.1` or `example.com` | Hostname or IP address |

What to expect after enabling it:
- if the runtime has the required capability, ICMP monitor creation is available and incident diagnostics can include ICMP reachability hints
- if the runtime does not have the required capability, startup still succeeds, the UI warns that ICMP is unavailable, and ICMP monitor creation is blocked until the capability is granted

---

### 1.3 Add a Heartbeat monitor (cron jobs and background workers)

Heartbeat monitoring verifies that scheduled tasks actually ran — not just that a service is reachable. Your job signals Ogoune when it finishes successfully. If no signal arrives within the configured window, an incident is created.

#### When to use it

- Nightly database backups
- Hourly data-import jobs
- Any cron job or background worker you need to confirm actually executed

#### Create the monitor

Open **Monitors → + Add Monitor**, then:

| Field | Example | Notes |
|---|---|---|
| **Type** | `Heartbeat` | Push-based check |
| **Name** | `Nightly backup` | Human-friendly label |
| **Ping Interval** | `300` | Expected seconds between pings (60–86400) |
| **Grace Period** | `60` | Extra seconds after deadline before alerting (60–3600) |

After saving, a **Ping URL** is generated. It looks like:

```
https://your-ogoune-host/ping/550e8400-e29b-41d4-a716-446655440000
```

Copy it — you will need it in the next step.

#### Add the ping call to your script

Append a ping request at the end of your job, **after the main work completes successfully**:

```bash
#!/usr/bin/env bash
set -euo pipefail

# your job
./run-backup.sh

# notify Ogoune on success
curl -fsS "https://your-ogoune-host/ping/<slug>" >/dev/null
```

Because the script uses `set -euo pipefail`, a failure in `./run-backup.sh` aborts before the `curl` line — so Ogoune is only pinged on success.

#### Schedule the job (cron example)

```cron
*/5 * * * * /opt/jobs/backup.sh
```

#### Grace period guidance

Set the grace period to roughly 10–20% of the interval. For a job that runs every 5 minutes (300 s), a 60 s grace is appropriate. For a 24-hour job, 1800 s (30 minutes) is reasonable.

#### What to expect

| Event | Result |
|---|---|
| First ping received | Monitor transitions `waiting → up` |
| No ping within interval + grace | Detector creates a missed-heartbeat incident |
| Ping received while down | Incident resolved automatically |
| Monitor is paused | Ping returns 403 — no state change |

---

## Step 2 — Configure notifications

Ogoune supports two notification channels out of the box: **SMTP (email)** and **Webhook** (Slack, Google Chat, Teams, Discord, or any HTTP endpoint).

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

Click **Test** before saving. Ogoune sends a test email immediately. If it arrives, the config is correct.

---

### Webhook — Slack, Google Chat, Teams, Discord

Webhooks work by sending an HTTP POST to a URL when an incident occurs. Every service listed below provides a webhook URL you paste into Ogoune.

#### Slack

1. Go to [api.slack.com/apps](https://api.slack.com/apps) → **Create New App** → **From scratch**
2. Name it `Ogoune`, choose your workspace
3. Click **Incoming Webhooks** → toggle **On**
4. Click **Add New Webhook to Workspace** → pick a channel → **Allow**
5. Copy the webhook URL — looks like: `https://hooks.slack.com/services/T.../B.../...`

In Ogoune:

| Field | Value |
|---|---|
| **Name** | `Slack — #alerts` |
| **Type** | `Webhook` |
| **URL** | the URL copied from Slack |
| **Enabled by default** | ✅ Toggle on |

#### Google Chat

1. Open the Google Chat space where you want alerts
2. Click the space name → **Apps & integrations** → **Add webhooks**
3. Name it `Ogoune` → **Save** → copy the URL

Paste the URL in Ogoune exactly as you did for Slack.

#### Microsoft Teams

1. In Teams, right-click a channel → **Manage channel** → **Connectors**
2. Search for **Incoming Webhook** → **Add** → **Add** again
3. Name it `Ogoune` → upload an icon (optional) → **Create**
4. Copy the webhook URL

#### Discord

1. In Discord, open channel settings → **Integrations** → **Webhooks**
2. Click **New Webhook** → name it → copy the URL
3. **Important:** append `/slack` to the URL — Discord's Slack-compatible webhook endpoint

```
https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN/slack
```

Paste this modified URL in Ogoune.

#### Custom HTTP endpoint

Any endpoint that accepts a POST with a JSON body works. Ogoune sends:

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

ICMP:
- `ENABLE_ICMP=false` by default
- set `ENABLE_ICMP=true` to enable ICMP monitor creation and ICMP-backed incident diagnostics
- if the runtime lacks the required raw-socket capability, Ogoune continues to start and reports ICMP as unavailable instead of failing startup

```env
# ── Database ────────────────────────────────────────────────────────────────
DB_DRIVER=sqlite              # sqlite | postgres
SQLITE_PATH=/data/ogoune.db  # Path to SQLite file (Community Edition)
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
ENABLE_ICMP=false             # Optional ICMP monitoring

# ── Security ─────────────────────────────────────────────────────────────────
JWT_SECRET=                   # Required — use `openssl rand -hex 32` to generate
ADMIN_EMAIL=admin@ogoune.test  # Default admin account email
```

---

## Troubleshooting

### ICMP shows as unavailable

Symptoms:
- The monitor form warns that ICMP is unavailable
- `GET /api/system/capabilities` returns `icmp.capability_available=false`
- Creating an ICMP monitor returns `422`

Checks:
- Confirm `.env` contains `ENABLE_ICMP=true`
- If running in Docker, grant the container the network capability required for raw ICMP sockets on your platform
- Restart Ogoune after changing environment or container capabilities

Expected behavior:
- Ogoune startup still succeeds even when ICMP capability is missing
- Existing HTTP, TCP, and DNS monitoring continues unchanged

### Container starts but I can't open the UI

```bash
# Check the container is running
docker ps | grep ogoune

# Check logs for errors
docker logs ogoune

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

**Step 3 — Check Ogoune logs:**
```bash
docker logs ogoune | grep "NOTIFICATION"
```
Look for `[WARNING] No notification channels configured` — this means the monitor has no channel assigned and no default is set.

**Step 4 — Check the incident event steps:**
Open the incident detail page. The timeline shows whether an alert was sent. If `alert_sent` step is missing, no notification was dispatched.

---

### Gmail SMTP returns "authentication failed"

You must use an **App Password**, not your regular Gmail password. See [Step 2 — Gmail setup](#gmail-setup).

---

### SQLite database file not persisting between restarts

You must mount a volume. Without `-v ogoune_data:/data`, data is lost when the container stops.

```bash
# Correct
docker run -v ogoune_data:/data ogoune/community:latest

# Wrong — data lost on restart
docker run ogoune/community:latest
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

---

## Verifying page content with Keyword checks

### When to use a Keyword monitor vs an HTTP monitor

Use an **HTTP monitor** when you only need to know the server responds with a success status code (2xx/3xx). Use a **Keyword monitor** when the status code alone is not enough — for example, when a site returns HTTP 200 but the page contains an error message, a maintenance notice, or is missing expected content.

### `contains` vs `not_contains`

|Mode|When to use|Raises incident when…|
|---|---|---|
|`contains`|Expect a string to always be present (e.g., `"operational"`, `"Welcome"`)|The keyword is **absent** from the response body|
|`not_contains`|Expect a string to never appear (e.g., `"maintenance"`, `"error"`)|The keyword is **present** in the response body|

**Example — status page check (contains):**

```json
{
  "type": "keyword",
  "target": "https://status.example.com",
  "keyword": "operational",
  "keyword_mode": "contains"
}
```

**Example — error detection (not_contains):**

```json
{
  "type": "keyword",
  "target": "https://app.example.com/health",
  "keyword": "error",
  "keyword_mode": "not_contains"
}
```

### 512 KB body limit

Ogoune reads at most **512 KB** of the response body for keyword inspection. Content beyond this limit is silently discarded. The `body_truncated` flag is set to `true` in incident diagnostics when truncation occurs. This cap prevents unbounded memory use on large responses (HTML pages, large JSON payloads).

For most status pages, health endpoints, and JSON APIs, 512 KB is more than sufficient.

### Case sensitivity

Keyword matching is **case-sensitive**. The keyword `"Operational"` will not match `"operational"`. Choose your keyword casing to match the exact text as it appears in the response body.

---

## Next steps

- [Architecture documentation](./ARCHITECTURE.md) — how Ogoune works under the hood
- [Contributing guidelines](./CONTRIBUTING.md) — how to contribute code or feedback
- [GitHub Discussions](https://github.com/denisakp/ogoune/discussions) — ask questions
- [GitHub Issues](https://github.com/denisakp/ogoune/issues) — report bugs

---

*Found something wrong or missing in this guide? [Open an issue](https://github.com/denisakp/ogoune/issues/new?labels=documentation) — doc fixes are always welcome.*