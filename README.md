<div align="right">
  <img src="./static/logo.png" width="40" alt="Ogoune" />
</div>

# Ogoune

**Uptime monitoring that confirms before it cries wolf.**

![License: Apache 2.0 (core)](https://img.shields.io/badge/license-Apache_2.0-blue) ![License: EE](https://img.shields.io/badge/internal%2Fee-LicenseRef--Ogoune--EE-orange)
![Version](https://img.shields.io/badge/version-v1.0.0--beta-yellow)
![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8)
![Vue](https://img.shields.io/badge/vue-3.x-4FC08D)
![Docker](https://img.shields.io/badge/docker-ready-2496ED)
[![GitHub Stars](https://img.shields.io/github/stars/denisakp/ogoune?style=flat)](https://github.com/denisakp/ogoune)
[![GitHub Actions](https://img.shields.io/badge/GitHub%20Actions-passing-brightgreen?logo=github)](https://github.com/denisakp/ogoune/actions)

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
| Password | `password` |

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
| Bulk import / export | Onboard or migrate hundreds of monitors via YAML — see [`docs/import/`](./docs/import/) |

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

Everything is configured through environment variables — start from
[`.env.example`](./.env.example). Only `APP_SECRET_KEY` (64-char hex,
`openssl rand -hex 32`) and `JWT_SECRET` are mandatory before production.

Full setup walkthrough (Community, full stack, source), the env-var reference,
ICMP opt-in, heartbeat integration, and troubleshooting live in
**[QUICKSTART.md](./QUICKSTART.md)**. For the `SSL_PROVIDER` matrix, magic-link
reset flow, and session-revocation semantics see
[`docs/runbooks/settings-env.md`](./docs/runbooks/settings-env.md).

---

## Roadmap

See [ROADMAP.md](./ROADMAP.md) for the full roadmap.

**Shipped since v1.0:** Keyword checks, Prometheus metrics, protocol-aware monitors,
bulk import/export, heartbeat monitoring, API v1, credential encryption.

---

## Observability

Ogoune exposes a Prometheus-compatible `GET /metrics` endpoint (opt-in via
`ENABLE_METRICS=true`, optional bearer token). Metric catalogue, access modes,
and scrape config: [`docs/observability/prometheus.md`](./docs/observability/prometheus.md).

---

## Contributing

Ogoune welcomes contributions. See [CONTRIBUTING.md](./CONTRIBUTING.md).

All contributors must sign the **CLA** — the bot handles this automatically on your first PR.

- **Report bugs** → [GitHub Issues](https://github.com/denisakp/ogoune/issues)
- **Request features** → [GitHub Discussions](https://github.com/denisakp/ogoune/discussions)
- **Good first issues** → [`good first issue`](https://github.com/denisakp/ogoune/labels/good%20first%20issue)

---

## Licence

Ogoune uses an Open Core dual-licensing model:

- **Core** (everything outside `internal/ee/`): **Apache License 2.0** — see [LICENSE](./LICENSE). Free, self-hostable, modifiable, and redistributable under standard permissive terms.
- **Enterprise Edition**: any file under `internal/ee/` or carrying the SPDX identifier `LicenseRef-Ogoune-EE` is governed by a separate commercial source-available licence — see [LICENSE.ee](./LICENSE.ee). The source is visible for evaluation, development, testing, and contribution, but production use requires a commercial licence (`hello@ogoune.com`).

Contributions to either scope require accepting the Contributor License Agreement — see [cla.md](./cla.md). The CLA bot prompts you on your first pull request.

**Prior licensing**: Ogoune's core was previously licensed under **AGPL v3**. The relicensing to Apache 2.0 predates any tagged release — no version was ever published under AGPL. Any copy obtained under AGPL remains governed by AGPL; the open-core model above governs the current source tree and all releases from `v1.0.0-beta` onward.


---

<div align="center">

Built with ❤️ by [Denis Yaovi](https://github.com/denisakp)

**[⬆ Back to top](#ogoune)**

</div>