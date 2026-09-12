# Roadmap

This document is public and intentionally transparent. It shows what we have built, what we are working on, and where 
we are going.

> **Last updated:** September 2026 — v1.0.0-beta.4 · H3 observability track re-ordered per the Execution Directive
> **Strategic context:** see [`BUSINESS-MODEL.md`](./BUSINESS-MODEL.md) for our open-core philosophy.

---

## Our model

Ogoune follows an **Open Core** model:

- The **Community Edition** is free, self-hosted, and open source under **Apache License 2.0**. Everything you need to monitor your infrastructure, detect failures before users do, and explain *why* they happened belongs here. Nobody should have to pay for production-grade observability.
- The **Enterprise Edition** adds features that only make sense in a multi-tenant Cloud context — team management, SSO, multi-org isolation, SOC 2 audit logs, dedicated support. This code lives in `internal/ee/` and is covered by a separate commercial source-available licence (`LicenseRef-Ogoune-EE`).

**We will never degrade the Community Edition to force upgrades.** If a feature ever moves between editions, it can only move *into* CE — never *out* of it.

---

## Licensing

| Edition | Licence | Code location |
|---|---|---|
| Community Edition | **Apache License 2.0** (see [`LICENSE`](./LICENSE)) | All files except `internal/ee/` |
| Enterprise Edition | **Commercial source-available** — `LicenseRef-Ogoune-EE` (see [`LICENSE.ee`](./LICENSE.ee)) | `internal/ee/` and any file carrying the SPDX identifier `LicenseRef-Ogoune-EE` |

The Apache 2.0 licence on the core lets you use, modify, deploy, and fork Ogoune Community Edition under standard permissive terms — no copyleft obligation. For Enterprise features in production, contact us for a commercial licence (`hello@ogoune.com`).

**Prior licensing.** Ogoune's core was previously licensed under AGPL v3; the relicensing to Apache 2.0 predates any tagged release. Any copy obtained under AGPL remains AGPL in perpetuity; the dual model above governs the current tree and all releases from `v1.0.0-beta` onward.

### Contributing

All contributors must sign our **CLA (Contributor Licence Agreement)** before their pull request can be merged. The current version is v1.1 (see [`cla.md`](./cla.md)). The CLA authorises us to licence contributions under any OSI-approved open source licence (currently Apache 2.0 on the core) and under any proprietary or commercial licence (currently `LicenseRef-Ogoune-EE` on `internal/ee/`). The CLA bot handles signature automatically on your first PR.

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
- [x] YAML bulk import/export for resources
- [x] Configurable dashboards — widget layout per view, scoped by tag/component/type/manual selection

---

## H2 — Shipped

Community Edition — expanding monitoring coverage, observability, and security.

### New monitor types

- [x] **Ping / ICMP** — check network reachability of any host
- [x] **Heartbeat / Push** — detect silent failures in cron jobs and background workers
- [x] **Keyword / content check** — verify a page contains an expected string, not just a 200 OK
- [x] **Protocol-aware checks** — application-layer handshake verification for Redis, MongoDB, FTP, and SSH.
  Confirms the service responds correctly at the protocol level, not just that the port is open.
  No credentials required. Extensible architecture — RabbitMQ, Kafka, and others are community contribution candidates.
- [ ] ~~**IMAP / SMTP**~~ — deferred indefinitely. Low demand, complex auth dependency. Webhook covers
  the same alerting use cases.

### Observability

- [x] **Prometheus metrics endpoint** — `GET /metrics` exposing runtime Go metrics and business metrics
  (resource status, check latency, incident counts, uptime ratios) for Grafana integration.
  Opt-in via `ENABLE_METRICS=true`. Optional bearer token auth via `METRICS_TOKEN`.

### API & Security

- [x] **Public API v1** — versioned REST API with OpenAPI spec
- [x] **Credential encryption (AES-256-GCM)** — SMTP passwords and webhook tokens encrypted at rest
- [x] **Server-side search endpoint** — `GET /api/v1/search`, ranked cross-entity search
  (monitors, incidents, app pages) powering the ⌘K command palette beyond the client-side window
- [x] **Announcements** — operator-authored, dismissible in-app banners (info/warning/success/error)

---

## H3 — Planned (Q4 2026)

**Community Edition focus quarter.** We're shipping the differentiators that make Ogoune stand apart: kernel-level observability via the agent, configurable escalation, and richer integrations. The vast majority of H3 is CE.

### Monitoring extensions

- [x] **Protocol-aware checks — auth variants** — Redis AUTH, MySQL, PostgreSQL with encrypted credentials.
  Depends on credential encryption (AES-256-GCM) shipped in H2. **Community Edition.**
- [x] **Protocol-aware checks — broker support** — RabbitMQ (AMQP handshake), Kafka (Metadata Request).
  Community contribution candidates — architecture is extensible from H2. **Community Edition.**
- [ ] **Database health checks** — service-health signals collected on the connection the Postgres and
  MySQL protocol checks already open: connection pool saturation, replication lag, longest running
  query. No server extension, no server restart, no extra credential. **Community Edition.**
  Extension naturelle des protocol-aware checks.
- [ ] **Multi-location checks (self-hosted)** — deploy Ogoune workers in multiple regions of your own
  infrastructure. Require N of M regions to fail before alerting. **Community Edition.** Cf. EE for the
  managed multi-region service.

### Agent device monitoring — the killer feature

- [x] **Lightweight agent (Go)** — **Linux** device monitoring agent (amd64 + arm64),
  shipped ahead of schedule in the v1.0.0-beta (specs 079-083). CPU, memory, per-mount disk, and network
  metrics streamed over a WebSocket via a reverse tunnel (agent initiates outbound connection), no
  inbound ports required. Single language across backend + agent (shared domain types via
  `pkg/agentwire`). Ships as a distroless container image + static release binaries, with a systemd
  unit and agent-down alerting. **Linux only, deliberately** — the packaging is systemd-based, and the
  servers this targets are Linux. macOS would need a launchd story and Windows a service wrapper, and
  neither is testable for us today. `nebula/self-host/agent.md` is the source of truth for platform
  support. Kernel-event capture (OOM kills, segfaults) ships with it — see Flash Correlation below.
- [x] **Flash Correlation (host metrics + kernel events)** — shipped in v1.0.0-beta.5. When a
  synthetic check fails (e.g. HTTP 502), the incident carries what the host was doing at that moment:
  the metrics the agent already streams, plus OOM kills and segfaults read from `/dev/kmsg` and cgroup
  v2 `memory.events`. The alert and the incident page open with one sentence: *"HTTP check failed:
  502 at 14:03:00. The kernel OOM-killed postgres (pid 4711) on web-01 at 14:02:47, 13 seconds
  earlier."* A time window and nothing else — it says *earlier*, never *because*. Since the next
  release an incident also records the machine it happened on, so moving a monitor does not rewrite it.
  **Community Edition.** No competitor offers this combination in open source today.
  > **How it works.** The agent tunnel is outbound and write-only by design: the agent pushes, the
  > backend never queries it. Correlation is therefore a backend-side join over data the agent has
  > already sent, not a request/response round trip. This needs no eBPF — plain file reads on the
  > agent side are enough for the flagship claim above.

### Status pages & branding

- [x] **Custom domain status page** — serve your status page on `status.yourdomain.com`. **Community Edition.**
- [x] **Status page branding** — custom logo (light/dark), primary color, and theme overrides on the
  public status page. **Community Edition.** Removing the "Powered by Ogoune" attribution is *not* part
  of this — it is an Enterprise lever (see "White-label — strict" below); Community Edition always
  keeps the attribution.
- [x] **Live incident updates** — editorial updates posted during an active incident (Investigating → Identified → Monitoring → Resolved), shown live on the public status page alongside the auto-detected uptime. Closes the loop between automatic detection and the planned Postmortem editor. **Community Edition.** Optional follow-ups not yet shipped: scheduled maintenance announcements, manual component degradation override, subscriber notifications (email/RSS).

### Reporting

- [x] **Scheduled reports — Community** — monthly health report (fixed schedule, first day of the month),
  shipped ahead of schedule in v1.0.0-beta. Covers all resources. Sent via configured SMTP channel.
  Toggle on/off in settings. No configuration required.
- **Non-goal: correlation reports.** No aggregate of incidents "caused by" a kernel event. Flash
  Correlation is a co-occurrence inside a time window; a report titled "caused by" promotes it to
  causality, summed and numbered, in a document that gets forwarded. May be revisited under
  Scheduled reports — Enterprise, and only as "incidents with X observed in the window".

### Alerting & integrations

- [x] **Escalation policies — Community** — native multi-step alert ladders. No PagerDuty required.
  Step N → wait X minutes → step N+1, with different channels per step. **Community Edition.**
- [ ] **Cloud integrations** — Vercel, Cloudflare, Coolify, Azure. OAuth flow + auto-discovery of resources.
  Just integrations API-level, no architectural lock. **Community Edition.**

### Security & encryption

- [ ] **Vault-backed credential encryption — Community** — basic config for HashiCorp Vault, AWS KMS,
  Azure Key Vault, GCP Secret Manager as the key provider. Replaces `APP_SECRET_KEY` with external
  key management. **Community Edition.** EE adds FIPS/HIPAA certified configs.

### Toolbox & utilities

- [x] **Toolbox** — one-off network checks. DNS lookup, Port scanner, SSL checker, WHOIS lookup. Manual
  triggers, no scheduling. CTA "Save as monitor" from results. **Community Edition.**

### Compliance (basic)

- [ ] **GDPR compliance tools — basic** — user data export (full JSON dump), data deletion request flow,
  privacy dashboard in Account Settings. **Community Edition.** EE adds DPA template, automated workflows.

---

## H3 — Enterprise Edition (Q4 2026 / Q1 2027)

EE is **smaller now**. Only features that genuinely require multi-tenancy, compliance certification, or 
the managed service itself live here. Everything else is in CE.

### Cloud foundation

- [ ] **Multi-tenancy** — organisation isolation, data separation. Required for Cloud and shared
  deployments. **Architecturally distinct from CE single-tenant.**
- [ ] **Onboarding + autonomous signup (Cloud)** — sign up → workspace setup wizard → plan choice →
  first monitor. **Cloud only** (the CE local install is the onboarding for CE).
- [ ] **Billing (Stripe)** — plans, quotas, invoicing, payment methods, proration on upgrade/downgrade.
  **Cloud only.**

### Team & access

- [ ] **Team management** — roles (Owner, Admin, Member, Viewer). Invitation flow with email. Permission
  matrix per role. **Requires multi-tenancy.**
- [ ] **SSO / SAML** — integration with Okta, Auth0, Azure AD, Google Workspace, generic SAML 2.0.
  Attribute mapping, JIT provisioning. **Compliance + complexity.**

### Compliance & audit

- [ ] **Audit logs — SOC 2 readiness** — full audit trail of all actions (user, action, resource, IP, 
  outcome). Filterable, exportable CSV. Required for SOC 2 certification process. **Cert investment.**
- [ ] **GDPR compliance tools — advanced** — DPA template generation, automated workflows, configurable
  retention per data category. **Builds on CE basic GDPR tools.**
- [ ] **Vault-backed credential encryption — certified** — FIPS 140-2 and HIPAA validated configurations
  for regulated industries. **Builds on CE Vault basic.**

### Reporting (Enterprise)

- [ ] **Scheduled reports — Enterprise** — configurable frequency (daily / weekly / custom cron), 
  filterable scope (by tag or component), multiple recipients. Built on the Community report engine.
- [ ] **SLA reports** — exportable PDF / CSV, branded with customer logo and colors. Required for
  audit-driven business contexts.

### Status pages (Enterprise tier)

- [ ] **White-label — strict** — remove "Powered by Ogoune" from all customer-facing surfaces
  (status page, email branding, generated PDFs). **Commercial lock, not architectural.**
- [ ] **Multi-status-page** — operate multiple branded status pages under one org (e.g. one per customer
  for agencies/MSPs). **Requires multi-tenancy.**

### Alerting (Enterprise)

- [ ] **PagerDuty / OpsGenie** — integration channels on the existing webhook pattern. Enterprise
  integration, **built on the first real customer request** — no internal owner, no date. Escalation
  policies (Community) already cover the on-call ladder without either.
- [ ] **Escalation policies — advanced** — on-call rotation schedules, override windows, complex routing
  rules with conditions. Builds on CE basic escalation. **Multi-user dependency.**

### Cloud service offering

- [ ] **Ogoune Cloud — hosted** — fully managed service in our global regions (US, EU, AP). 
  Auto-scaling, multi-region failover, 99.99% SLA contractual. **The service itself, not just a feature.**
- [ ] **Multi-location checks — managed regions** — verify from Ogoune-operated regions before alerting.
  Different from CE multi-location (your own infrastructure). **Infrastructure opex on our side.**
- [ ] **Dedicated support + SLA** — portal, Slack channel, email, 24h business-days response. SOC 2 
  Type II report, security questionnaires, DPA, etc. **Human service.**

### Mobile companion

- [ ] **Mobile app companion (Cloud only)** — read-only incident consumption + acknowledgement. PWA first,
  native iOS/Android later if traction justifies. **Cloud only** — no architecturally clean way to
  connect a mobile app to arbitrary self-hosted instances (would require user to expose Ogoune publicly,
  defeats the security model).

---

## H4 — Long-term vision (2027+)

### Community Edition

- [ ] **Flash Correlation — eBPF depth** — the signals that genuinely require kernel instrumentation:
  syscall latency, TCP retransmits, packet drops. Via `cilium/ebpf`. Deferred out of H3 deliberately:
  the differentiator is the correlation, not the capture, and Flash Correlation ships without it.
  eBPF carries a real cost (BTF/CO-RE kernel floor, `bpf2go` toolchain, privileged load, Linux-only
  dev environment) for the least legible signals. **CE.**
  > **Note:** A Zig rewrite for sub-3MB footprint was previously planned and dropped. eBPF tooling is
  > mature in Go (Cilium, Datadog, Parca, Pixie all ship Go+eBPF in prod), Zig is pre-1.0 with weak
  > Windows/macOS coverage, and footprint <30MB is not a real adoption blocker. If a future profile
  > proves footprint matters commercially, Rust would be preferred over Zig for eBPF ecosystem maturity.
- [ ] **Slow query analysis (via agent)** — per-query performance and regression detection on
  Postgres/MySQL, read locally by the agent (`performance_schema`, log files) rather than over a
  network credential. Split out of the H3 "Database query performance" item. **CE, exploratory —
  APM boundary, re-arbitrate before committing.** Two reasons it is not in H3: `pg_stat_statements`
  requires `shared_preload_libraries` and therefore a **server restart**, which is fatal to trial
  adoption; and per-query analysis sits against the APM boundary this roadmap declares out of scope.
- [ ] **Log aggregation lite** — structured logs shipped by the host agent (kernel events, host state at
  failure moments). Not a full log search engine; a timeline view tied to incidents. **CE.**
- [ ] **AI / Predictive analytics** — detect anomalies before they cause incidents. Use the agent metrics
  + incident history. Run locally (no LLM API calls required). **CE.**
- [ ] **MCP integration** — conversational monitoring via Claude / ChatGPT / your local Llama. Read-only,
  query state and incidents via natural language. **CE.**
- [ ] **Phone call alerts** — for critical incidents needing to wake someone up. Twilio integration in CE,
  bring-your-own-credentials. **CE.**

### Enterprise Edition

- [ ] **AI / Predictive — advanced** — anomaly detection across multiple orgs, cross-customer patterns
  (with privacy-preserving aggregation). **EE.**
- [ ] **Phone call alerts — managed** — Twilio + Vonage included, no BYO credentials, billed per minute.
  **EE Cloud.**

---

## Explicitly out of scope

| Feature | Reason |
|---|---|
| **Real User Monitoring (RUM)** | Different product category. Sentry, PostHog, Datadog RUM territory. Scope creep mortal. |
| **Synthetic transaction monitoring** (multi-step scripted user journeys) | Massive scope (Playwright sandbox), different category from synthetic checks. |
| **Full distributed tracing** (Jaeger/Tempo replacement) | Massive infrastructure cost, different category. We accept OTLP and enrich, but don't store traces. |
| **Full log search engine** (Elastic/Loki replacement) | Different category. We ship lite log aggregation from the agent, no full-text search engine. |
| **Custom scripting language for checks** | Sandboxing nightmare, RCE attack surface. Use webhooks / heartbeats instead. |
| **APM SDKs** (Java/Python/Node auto-instrumentation) | Different category. Use OpenTelemetry SDKs + send to Ogoune via OTLP. |
| **Browser / synthetic monitoring** (visual diff, headless browser) | Different product scope. Use Checkly, Better Stack. |
| **SCIM / HRIS provisioning** | Needed only at 500+ users per org. Not before Q4 2027 at earliest. |
| **SMS notifications** | Low adoption — Webhook covers the same use cases. Phone call is the escalation channel. |
| **IMAP / SMTP checks** | Low demand — TCP check covers port-level verification; complex auth adds risk without proportional value. |
| **Digest / real-time notification batching** | Uptime alerts are time-critical — batching defeats the purpose. See Scheduled reports instead. |

---

## What changed between H2 and H3

The H3 roadmap was significantly rebalanced in May 2026 after a strategic review. **Twelve features previously labeled EE moved to CE**, reflecting the principle that features only belong in EE when they require multi-tenancy, compliance certification, or the managed service. The agent device monitoring + flash correlation in particular is now CE — it's the strongest differentiator of the product, and locking it behind EE would prevent the open-source flywheel from working.

EE is smaller, but more defensible: it's the managed service, multi-user infrastructure, and certified compliance — not features arbitrarily fenced off.

Two features were added: database query performance and Log aggregation lite (from the agent).
The database item was later split — health checks stayed in H3, per-query analysis moved to H4 — and
the agent shipped in Go, not Zig.
Three features were explicitly cut to avoid scope creep: RUM, Synthetic transactions, Full distributed tracing.

---

## How to influence this roadmap

- **[GitHub Discussions](https://github.com/denisakp/ogoune/discussions)** — propose features, explain why they matter
- **[GitHub Issues](https://github.com/denisakp/ogoune/issues)** — bugs and well-defined requests
- **Upvote** existing issues to signal demand
- **Open a PR** — see [CONTRIBUTING.md](./CONTRIBUTING.md)
- **Talk to us** — `hello@ogoune.com` for commercial / EE conversations

We read everything. We always explain when we decline something.

---

*Built by [@denisakp](https://github.com/denisakp). See [BUSINESS-MODEL.md](./BUSINESS-MODEL.md) for our open-core philosophy.*
