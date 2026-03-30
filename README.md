<div align="right">
  <img src="./static/ico.png" width="40" alt="Ogoune" />
</div>

# Ogoune

**Uptime monitoring that confirms before it cries wolf.**

![License](https://img.shields.io/badge/license-AGPL%20v3-blue)
![Version](https://img.shields.io/badge/version-v1.0.0-green)
![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8)
![Vue](https://img.shields.io/badge/vue-3.x-4FC08D)
![Docker](https://img.shields.io/badge/docker-ready-2496ED)
[![GitHub Stars](https://img.shields.io/github/stars/denisakp/ogoune?style=flat)](https://github.com/denisakp/ogoune)

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
  ogoune/community:latest
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
curl -o compose.yml https://raw.githubusercontent.com/denisakp/ogoune/main/docker-compose.community.yml
docker compose up -d
```

</details>

<details>
<summary>Or with the full stack (PostgreSQL + Redis)</summary>

```bash
git clone https://github.com/denisakp/ogoune.git
cd ogoune
cp .env.example .env
docker compose up -d
```

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
| Open source | AGPL v3 | MIT | ❌ | Apache |
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

## Roadmap

See [ROADMAP.md](./ROADMAP.md) for the full roadmap.

**Coming in H2:** Heartbeat/Push, Keyword checks, Prometheus metrics, IMAP/SMTP, Telegram, Digest
notifications, API v1, credential encryption.

**Coming in H3:** Toolbox, Enterprise Edition (multi-tenancy, SSO, billing, agent device monitoring).

---

## Contributing

Ogoune welcomes contributions. See [CONTRIBUTING.md](./CONTRIBUTING.md).

All contributors must sign the **CLA** — the bot handles this automatically on your first PR.

- **Report bugs** → [GitHub Issues](https://github.com/denisakp/ogoune/issues)
- **Request features** → [GitHub Discussions](https://github.com/denisakp/ogoune/discussions)
- **Good first issues** → [`good first issue`](https://github.com/denisakp/ogoune/labels/good%20first%20issue)

---

## Licence

Ogoune is licensed under **AGPL v3** — see [LICENSE](./LICENSE).

The `internal/ee/` directory contains Enterprise Edition features and is covered by a separate proprietary licence.
see [LICENSE_EE](./LICENSE_EE).

A valid licence key is required to use those features, whether self-hosted or via our Cloud. The rest of the codebase 
is free, open source, and self-hostable with no licence key required. See [ROADMAP.md](./ROADMAP.md) for the Open Core
model.


---

<div align="center">

Built with ❤️ by [denisakp](https://github.com/denisakp)

**[⬆ Back to top](#ogoune)**

</div>