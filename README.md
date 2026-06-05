<div align="right">
  <img src="./static/logo.png" width="40" alt="Ogoune" />
</div>

# Ogoune

**Uptime monitoring that confirms before it cries wolf.**

![License: Apache 2.0 (core)](https://img.shields.io/badge/license-Apache_2.0-blue) ![License: EE](https://img.shields.io/badge/internal%2Fee-LicenseRef--Ogoune--EE-orange)
![Version](https://img.shields.io/badge/version-v1.0.0-green)
![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8)
![Vue](https://img.shields.io/badge/vue-3.x-4FC08D)
![Docker](https://img.shields.io/badge/docker-ready-2496ED)
[![GitHub Stars](https://img.shields.io/github/stars/denisakp/ogoune?style=flat)](https://github.com/denisakp/ogoune)
[![GitHub Actions](https://img.shields.io/badge/GitHub%20Actions-passing-brightgreen?logo=github)](https://github.com/denisakp/ogoune/actions)
[![GitLab CI/CD](https://img.shields.io/badge/GitLab%20CI-ready-orange?logo=gitlab)](./specs/021-gitlab-ci-workflows/quickstart.md)

Ogoune monitors your websites, APIs, and services. When something goes down, it **verifies the failure** before 
alerting you. No more 3am pages for a 2-second network blip.

> Most monitoring tools alert on the first failed check. Ogoune confirms failures before creating an incident.
> Every alert is real.

<img src="./static/dashboard.png" alt="Ogoune Dashboard" width="100%" style="border-radius: 8px; margin-top: 16px;" />

---

## Get started in 10 seconds

```bash
docker run -d \
  -p 8080:8080 \
  -v ogoune:/data \
  --name ogoune \
  ghcr.io/denisakp/ogoune:latest
```

Open **http://localhost:8080** and log in:

| | |
|---|---|
| Email | `admin@ogoune.test` |
| Password | `ogu3n3@rd` |

No PostgreSQL. No Redis. No reverse proxy. One container.

<details>
<summary>Or with Docker Compose</summary>

```bash
git clone https://github.com/denisakp/ogoune.git
cd ogoune
cp .env.example .env   # set JWT_SECRET and APP_SECRET_KEY at minimum (see command below)
docker compose up -d
```

Defaults to **SQLite + in-process scheduler** — no external dependencies.

Generate `APP_SECRET_KEY` (required, 64-char hex):

```bash
openssl rand -hex 32
```

To switch to the full stack (PostgreSQL + Redis), add these to your `.env`:

```env
COMPOSE_PROFILES=full
DB_DRIVER=postgres
DATABASE_URL=postgres://ogoune:ogoune@postgres:5432/ogoune?sslmode=disable
SCHEDULER_MODE=asynq
REDIS_URL=redis://redis:6379
```

Add `ENTERPRISE_LICENSE_KEY=<your-key>` to activate Enterprise Edition features.

</details>

---

## Why Ogoune

**The false positive problem.** Most open-source monitors alert the moment a single check fails. Network hiccups, DNS
timeouts, rolling deploys, they all trigger alerts. After a few weeks, you stop reading them. Then you miss the real
outage.

Ogoune solves this by **confirming every failure** before alerting:

```
Check 1 → DOWN  →  waiting 30s, not alerting yet...
Check 2 → DOWN  →  confirmed. incident created. alert sent. ✓

Check 1 → DOWN  →  waiting 30s...
Check 2 → UP    →  false positive. no incident. no noise. ✓
```

Configure per monitor:

```
confirmation_checks: 2     # failures before alerting
confirmation_interval: 30s # gap between confirmation checks
```

Set `confirmation_checks: 1` to restore immediate alerts.

---

## What you get

| | |
|---|---|
| HTTP / HTTPS checks | Monitor websites and APIs |
| TCP port checks | Monitor any service port |
| DNS checks | Verify DNS resolution |
| ICMP ping checks | Optional host reachability monitoring when the runtime has raw-socket capability |
| Heartbeat / Push monitoring | Verify cron jobs and background workers actually ran |
| Keyword / content check monitor | Verify response body contains (or does not contain) a string — catches HTTP 200s with degraded content |
| Protocol-aware monitors | Redis (PING + optional `AUTH`), MongoDB (BSON hello), FTP, SSH, MySQL & PostgreSQL (TCP fallback or authenticated handshake), RabbitMQ (AMQP 0-9-1 `connection.start` handshake), Kafka (Metadata Request v1 against a comma-separated bootstrap broker list with sequential failover). TLS is auto-detected from the target URL (`rediss://`, `?tls=true`, `sslmode=require`). Credentials are encrypted at rest with AES-256-GCM. |
| SSL expiry warnings | Get notified before certs expire |
| Domain expiry warnings | Get notified before domains expire |
| Confirmation window | N consecutive failures before alerting |
| Flap detection | Suppress alerts for unstable resources |
| Alert grouping | One notification for simultaneous component failures |
| Incidents & timeline | Full lifecycle with rich diagnostics |
| SMTP notifications | Email alerts |
| Webhook notifications | Slack, Google Chat, Teams, Discord, any HTTP endpoint |
| Status page | Public page for your customers |
| Maintenance windows | One-time and recurring (cron) |
| 2FA | TOTP-based two-factor auth |
| Tags & components | Organize your monitors |
| API keys | Programmatic access (`read` / `read_write`) |
| Uptime statistics | 2h / 24h / 7d / 30d aggregates |

---

## Zero dependencies

The Community Edition runs on **embedded SQLite** with an **in-process scheduler**.

```env
DB_DRIVER=sqlite               # Community (default)
DB_DRIVER=postgres             # Production
SCHEDULER_DRIVER=timingwheel   # Community (default)
SCHEDULER_DRIVER=asynq         # Production (requires Redis)
```

---

## Comparison

| | Ogoune | Uptime Kuma | UptimeRobot | Prometheus + Alertmanager |
|---|---|---|---|---|
| Self-hosted | ✅ | ✅ | ❌ | ✅ |
| Zero dependencies | ✅ | ✅ | — | ❌ |
| False positive protection | ✅ | ❌ | ❌ | Manual |
| Flap detection | ✅ | ❌ | ❌ | Manual |
| SSL + domain expiry | ✅ | ✅ / ❌ | Paid | Manual |
| DNS monitoring | ✅ | ✅ | Paid | Manual |
| Open source | Apache 2.0 | MIT | ❌ | Apache 2.0 |
| Go backend | ✅ | ❌ | — | ✅ |
| Setup complexity | Low | Low | None | Very high |

---

## Configuration

All settings via environment variables. See [`.env.example`](./.env.example).

```env
# Database
DB_DRIVER=sqlite
SQLITE_PATH=/data/ogoune.db

# Scheduler
SCHEDULER_DRIVER=timingwheel

# Security — generate before production
JWT_SECRET=change-me
APP_SECRET_KEY=change-me    # openssl rand -hex 32

# Open Core
ENTERPRISE_LICENSE_KEY=     # leave empty for Community Edition

# Settings (spec 059)
APP_BASE_URL=https://status.example.com  # used in 2FA reset magic-link emails
SSL_PROVIDER=external                    # letsencrypt | external | disabled
```

See [`docs/runbooks/settings-env.md`](./docs/runbooks/settings-env.md) for the
`SSL_PROVIDER` UI matrix, the magic-link reset flow, and session-revocation
semantics.

### Generate `APP_SECRET_KEY`

`APP_SECRET_KEY` is mandatory and must be a 64-character hex string.

```bash
# Generate a key value
openssl rand -hex 32

# Optional: write it directly to .env
echo "APP_SECRET_KEY=$(openssl rand -hex 32)" >> .env
```

If you export it in your shell instead of `.env`, verify length:

```bash
echo -n "$APP_SECRET_KEY" | wc -c
```

### Optional ICMP monitoring

ICMP monitoring is opt-in because raw ICMP sockets depend on host/container capabilities that are not universally available by default.

Enable it only when you want ping-based monitoring or ICMP-backed network diagnostics for incidents:

```env
ENABLE_ICMP=true
```

When `ENABLE_ICMP=false` (default), Ogoune keeps HTTP, TCP, and DNS behavior unchanged and skips ICMP-specific checks and diagnostics.

When `ENABLE_ICMP=true`, Ogoune starts normally in both cases:

- if the runtime has the required capability, ICMP monitor creation and ICMP-backed diagnostics are available
- if the runtime does not have the required capability, startup continues, the UI/API report ICMP as unavailable, and ICMP monitor creation is rejected until capability is granted

---

### Heartbeat / Push monitoring

Heartbeat monitoring lets you verify that cron jobs and background workers actually ran. Instead of Ogoune polling a target, your job calls Ogoune at the end of a successful run.

```bash
# At the end of your script, ping Ogoune
curl -fsS "https://your-ogoune-host/ping/<slug>" >/dev/null
```

**How it works:**
- Create a Heartbeat monitor with an interval and grace period.
- Copy the generated ping URL and add it to your script.
- Ogoune waits. If no ping arrives within interval + grace seconds, an incident is created.
- A recovery ping resolves the incident automatically.

**No authentication required.** The slug itself is the token (UUID v4, unguessable). A per-slug rate limit of 100 requests/min applies.

See [QUICKSTART.md](./QUICKSTART.md) for step-by-step integration.

---

## Roadmap

See [ROADMAP.md](./ROADMAP.md) for the full roadmap.

**Coming in H2:** Keyword checks, Prometheus metrics, IMAP/SMTP, Telegram, Digest
notifications, API v1, credential encryption.

**Coming in H3:** Toolbox, Enterprise Edition (multi-tenancy, SSO, billing, agent device monitoring).

---

## Prometheus Metrics Endpoint

Ogoune exposes a standard Prometheus-compatible `GET /metrics` endpoint for observability.

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_METRICS` | `false` | Set to `true` to enable the `/metrics` endpoint |
| `METRICS_TOKEN` | _(empty)_ | Optional bearer token. When set, requests must include `Authorization: Bearer <token>`. Leave empty only on private/firewalled networks (a startup warning is logged). |

### Access modes

| `ENABLE_METRICS` | `METRICS_TOKEN` | Result |
|-----------------|-----------------|--------|
| `false` | any | `404 Not Found` — route not registered |
| `true` | _(empty)_ | `200 OK` — unauthenticated (startup warning logged) |
| `true` | `<token>` | `200 OK` with correct `Authorization: Bearer <token>` header; `401` otherwise |

### Available metrics

| Metric | Type | Description |
|--------|------|-------------|
| `go_goroutines` | Gauge | Active goroutines |
| `go_memstats_heap_alloc_bytes` | Gauge | Heap memory in use |
| `go_gc_duration_seconds` | Summary | GC pause durations |
| `ogoune_resource_up` | Gauge | `1`=up, `0`=down per resource |
| `ogoune_resource_status` | Gauge | `0`=unknown `1`=up `2`=down `3`=paused |
| `ogoune_check_duration_seconds` | Histogram | Check latency in seconds |
| `ogoune_checks_total` | Counter | Check executions by `status` label (`success`/`failure`/`timeout`) |
| `ogoune_incidents_total` | Gauge | All-time incident count per resource |
| `ogoune_incidents_active` | Gauge | Currently open incidents per resource |
| `ogoune_uptime_ratio` | Gauge | Uptime `0.0–1.0` for `window` label `24h`, `7d`, `30d` |

All `ogoune_*` metrics carry labels: `id`, `name`, `type`.

### Prometheus scrape config

Without authentication:
```yaml
scrape_configs:
  - job_name: ogoune
    scrape_interval: 30s
    static_configs:
      - targets: ["ogoune:8080"]
```

With bearer token:
```yaml
scrape_configs:
  - job_name: ogoune
    scrape_interval: 30s
    bearer_token: your-secret-token-here
    static_configs:
      - targets: ["ogoune:8080"]
```

---

## Contributing

Ogoune welcomes contributions. See [CONTRIBUTING.md](./CONTRIBUTING.md).

All contributors must sign the **CLA** — the bot handles this automatically on your first PR.

- **Report bugs** → [GitHub Issues](https://github.com/denisakp/ogoune/issues)
- **Request features** → [GitHub Discussions](https://github.com/denisakp/ogoune/discussions)
- **Good first issues** → [`good first issue`](https://github.com/denisakp/ogoune/labels/good%20first%20issue)

---

## Licence

Ogoune uses an Open Core dual-licensing model from v2.0.0 onward:

- **Core** (everything outside `internal/ee/`): **Apache License 2.0** — see [LICENSE](./LICENSE). Free, self-hostable, modifiable, and redistributable under standard permissive terms.
- **Enterprise Edition**: any file under `internal/ee/` or carrying the SPDX identifier `LicenseRef-Ogoune-EE` is governed by a separate commercial source-available licence — see [LICENSE.ee](./LICENSE.ee). The source is visible for evaluation, development, testing, and contribution, but production use requires a commercial licence (`hello@ogoune.com`).

Contributions to either scope require accepting the Contributor License Agreement — see [cla.md](./cla.md). The CLA bot prompts you on your first pull request.

**Past releases**: any Ogoune distribution made publicly available prior to v2.0.0 remains licensed under **AGPL v3 forever**. The dual-licensing change above applies only to commits and releases made from v2.0.0 onward. See the [migration announcement](https://github.com/denisakp/ogoune/discussions) for details.


---

<div align="center">

Built with ❤️ by [Denis Yaovi](https://github.com/denisakp)

**[⬆ Back to top](#ogoune)**

</div>