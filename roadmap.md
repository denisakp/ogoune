# Roadmap

This document is public and intentionally transparent. It shows what we have built, what we are working on, and where 
we are going.

> **Last updated:** March 2026 — v1.0.0

---

## Our model

Ogoune follows an **Open Core** model:

- The **Community Edition** is free, self-hosted, open source under **AGPL v3**. Everything you need to monitor your
infrastructure and get alerted belongs here. Nobody should have to pay to know if their service is down.
- The **Enterprise Edition** adds features that only make sense in a hosted, multi-tenant context — team management, SSO,
enterprise integrations, AI analytics. This code lives in `internal/ee/` and is covered by a separate commercial licence.

**We will never degrade the Community Edition to force upgrades.**

---

## Licensing

| Edition | Licence | Code location |
|---|---|---|
| Community Edition | **AGPL v3** | All directories except `internal/ee/` |
| Enterprise Edition | Proprietary (see `LICENSE_EE`) | `internal/ee/` |

The AGPL v3 licence means anyone running Ogoune as a service must publish their modifications. If you want to use Ogoune
commercially without sharing your modifications, contact us for a commercial licence.

### Contributing

All contributors must sign our **CLA (Contributor Licence Agreement)** before their pull request can be merged. This
allows us to license community contributions under AGPL v3 while retaining the ability to offer commercial licences.
The CLA bot handles this automatically on your first PR.

---

## v1.0.0 — March 2026

**Community Edition — stable, zero external dependencies.**

### Monitoring
- [x] HTTP / HTTPS checks
- [x] TCP port checks
- [x] DNS resolution checks
- [x] SSL certificate expiry alerts (J-30, J-14, J-7, J-1)
- [x] Domain expiry alerts via WHOIS
- [x] SSL and domain metadata enrichment

### Alerting
- [x] Confirmation window — N consecutive failures before alerting (no false positives)
- [x] Flap detection — suppress alerts for unstable resources
- [x] Alert cooldown — exactly one "down" alert per incident
- [x] Timed reminders — optional re-alerts while incident is active
- [x] Component-level alert grouping — one notification for simultaneous failures
- [x] Pending notification retry on startup

### Incidents
- [x] Automatic incident lifecycle (creation, resolution)
- [x] Rich diagnostics (timing breakdown, failure classification)
- [x] Human-readable cause messages
- [x] Event timeline

### Notifications
- [x] SMTP (email)
- [x] Webhook — Slack, Google Chat, Teams, Discord, any HTTP endpoint
- [x] `enabled_by_default` : one channel covers all monitors

### Status Page
- [x] Public status page (unauthenticated)
- [x] Component-level status aggregation
- [x] 90-day uptime bar per monitor
- [x] Dual entry point — deployable on `status.yourdomain.com`
- [x] Maintenance windows (one-time + cron, suppresses false positives during planned downtime)

### Infrastructure
- [x] SQLite embedded — zero external dependencies (Community Edition)
- [x] PostgreSQL + Redis / Asynq — full production stack
- [x] TimingWheel in-process scheduler — no Redis required
- [x] Auto-refreshing monitor detail page
- [x] API keys (`read` / `read_write` scopes)
- [x] Two-factor authentication (TOTP)
- [x] Maintenance windows (one-time + cron)
- [x] Components, Tags, Organization

---

## H2 — In progress

Community Edition — expanding monitoring coverage and hardening security.

### New monitor types
- [ ] **Ping / ICMP** — check network reachability of any host
- [ ] **Heartbeat / Push** — detect silent failures in cron jobs and background workers
- [ ] **Keyword / content check** — verify a page contains an expected string, not just a 200 OK
- [ ] **IMAP / SMTP** — verify mail servers are responding
- [ ] **Extended protocols** — MySQL, PostgreSQL, Redis, MongoDB, FTP, SSH port checks with pre-filled defaults

### Observability
- [ ] **Prometheus metrics endpoint** — `GET /metrics` for Grafana integration
- [ ] **OpenTelemetry** — app-level traces and runtime metrics (Phase 1)

### Notifications
- [ ] **Digest mode** — bundled hourly / daily / weekly summaries

### API & Security
- [ ] **Public API v1** — versioned REST API with OpenAPI spec
- [ ] **Credential encryption (AES-256-GCM)** — SMTP passwords and webhook tokens encrypted at rest

---

## H3 — Planned (Q4 2026)

### Community Edition
- [ ] **Toolbox** — DNS lookup, Port scanner, SSL checker
- [ ] **Alerting escalation** — native, without PagerDuty
- [ ] **Monitor dashboards** — custom filtered views by tag or component

### Enterprise Edition — foundation
- [ ] **Multi-tenancy** — organisation isolation, data separation
- [ ] **Onboarding + signup** — autonomous account creation
- [ ] **Billing (Stripe)** — plans, quotas, invoicing

### Enterprise Edition — enterprise
- [ ] **SSO / SAML**
- [ ] **Team management** — roles (Owner, Admin, Member, Viewer)
- [ ] **Audit logs** — SOC 2 readiness
- [ ] **SLA reports** — exportable PDF / CSV, branded
- [ ] **White-label status page**
- [ ] **Custom domain status page**
- [ ] **Escalation policies**
- [ ] **GDPR compliance tools**
- [ ] **Multi-location checks** — verify from multiple regions before alerting
- [ ] **PagerDuty / OpsGenie**
- [ ] **Cloud integrations** — Azure, Vercel, Cloudflare, Coolify
- [ ] **Agent device monitoring** — CPU, memory, disk via lightweight Go agent

---

## H4 — Long-term vision

- [ ] **AI / Predictive analytics** — detect anomalies before they cause incidents
- [ ] **MCP integration** — conversational monitoring via ChatGPT / Claude
- [ ] **Phone call alerts** — for incidents that need to wake someone up

---

## Not on the roadmap

| Feature | Reason |
|---|---|
| Browser / synthetic monitoring | Different product scope |
| APM / distributed tracing of your apps | Datadog / Sentry territory |
| SCIM / HRIS provisioning | Needed only at 500+ users per org |
| SMS notifications | Low adoption — Webhook covers the same use cases |

---

## How to influence this roadmap

- **[GitHub Discussions](https://github.com/denisakp/ogoune/discussions)** — propose features, explain why they matter
- **[GitHub Issues](https://github.com/denisakp/ogoune/issues)** — bugs and well-defined requests
- **Upvote** existing issues to signal demand
- **Open a PR** — see [CONTRIBUTING.md](./CONTRIBUTING.md)

We read everything. We always explain when we decline something.

---

*Built by [@denisakp](https://github.com/denisakp)*
