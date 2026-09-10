# Monitor types

Ogoune ships several check strategies. Each implements a common `CheckStrategy`
contract, so new types can be added without touching the scheduler or worker pool.

| Type | Checks | Key fields |
|---|---|---|
| **HTTP** | Status code (a `HEAD` in the 200–399 range is UP), response time, redirects | `target` (full URL) |
| **TCP** | Port reachability | `target` (`host:port`) |
| **DNS** | Record resolution (`LookupHost`) | `target` (hostname) |
| **ICMP** | Ping / host reachability (single echo probe) | `target` (host) |
| **Keyword** | Presence/absence of a string in the response body | `keyword`, `keyword_mode` |
| **Heartbeat** | Push-based — you ping Ogoune on a schedule | `heartbeat_interval`, `heartbeat_grace` |
| **Protocol** | Application-layer handshakes (Redis, MongoDB, FTP, SSH, MySQL, PostgreSQL, RabbitMQ, Kafka) | `protocol_type`, `protocol_port` |

Every monitor also takes the common fields `name`, `type`, `interval` (10–3600 s),
`timeout` (1–60 s), and optional `confirmation_checks` / `confirmation_interval`
(how many consecutive failures confirm an incident before alerting).

## Keyword

A `GET` request reads up to 512 KB of the body, then applies a **case-sensitive**
substring match. `keyword_mode` is `contains` (UP when found) or `not_contains`
(UP when absent).

```json
{ "type": "keyword", "target": "https://example.com/health",
  "keyword": "OK", "keyword_mode": "contains" }
```

## Heartbeat (push)

Instead of Ogoune reaching out, a heartbeat monitor waits for **your** job (cron,
backup, worker) to check in. Ogoune generates a `heartbeat_slug`; ping it from your
job:

```
POST /api/v1/heartbeat/ping/{slug}
```

The endpoint is public — no auth header — so a `curl` at the end of a script is
enough. Configure `heartbeat_interval` (expected seconds between pings) and
`heartbeat_grace` (extra slack before it's marked down).

::: tip
A brand-new heartbeat stays in a *waiting* state until its first ping arrives, so
it never alerts before your job has run once.
:::

## SSL & domain expiry

SSL certificate and domain (WHOIS) expiry are **not** a separate monitor type —
they're enrichment on your active **HTTP** monitors. A daily job populates each
resource's SSL issuer / expiry date and domain registrar / expiry date, then
alerts as the deadline approaches.

Thresholds are set per monitor via `expiry_alert_thresholds` (a comma-separated
list of days-remaining, each 1–365). The default is `30,14,7,1`:

```json
{ "type": "http", "target": "https://example.com",
  "expiry_alert_thresholds": "30,14,7,1" }
```

Status is derived from days remaining: **expired** (≤ 0), **critical** (≤ 7),
**warning** (≤ 30), otherwise **ok**.

## Protocol with authentication

A protocol monitor picks its handler from `protocol_type` and connects on
`protocol_port` (or the default: Redis 6379, MongoDB 27017, FTP 21, SSH 22,
MySQL 3306, PostgreSQL 5432, RabbitMQ 5672, Kafka 9092). TLS is inferred from the
target — `rediss://`, `?tls=true`, or `?sslmode=require`.

For Redis, MySQL, and PostgreSQL you can go beyond port reachability and verify a
real login. Create the monitor, then attach a credential:

```
POST /api/v1/resources/{id}/credentials
{ "username": "monitor", "password": "s3cret" }
```

Redis uses `AUTH` (the ACL `AUTH <user> <pass>` form when a username is set);
MySQL and PostgreSQL open an authenticated connection and ping it. An auth failure
is reported distinctly from an unreachable port.

::: warning
Credentials are encrypted at rest (AES-256-GCM) and the password is never returned
by the API — reads show a mask. Without a credential, MySQL/PostgreSQL monitors
fall back to a plain TCP reachability check.
:::

## Database health on Postgres and MySQL monitors

A protocol monitor pointed at PostgreSQL or MySQL already opens an authenticated
session to check the server answers. It now also asks that server how it is
doing, on the same connection, and shows the answer on the monitor's page:

| Signal | Needs a grant? |
|---|---|
| Active connections against the configured maximum | no |
| Age of the longest running query | **yes** |
| Replication lag | **yes** |

**Nothing is required on the monitored database.** No extension, no restart, no
configuration change, no extra credential field. A monitor that works today keeps
working and simply shows more.

The first row is the one that earns its place: it lets you watch connection
saturation climb *before* the database starts refusing connections, which is one
of the most common non-obvious causes of an outage behind a healthy-looking
service.

### The optional grant

The last two rows need a credential that can see other sessions. Without it they
are **absent** — not zero, not estimated. That distinction is deliberate: on
PostgreSQL a role without the grant still sees other sessions' rows with the
timing columns withheld, so a naive reading would report your monitor's own
session age as the database's longest running query. Plausible, wrong, and
impossible for you to spot. We would rather show nothing.

To unlock them:

```sql
-- PostgreSQL
GRANT pg_read_all_stats TO your_monitoring_role;
```

```sql
-- MySQL
GRANT PROCESS ON *.* TO 'your_monitoring_user'@'%';
```

Both are read-only and entirely optional. The monitor page tells you when a grant
would add something, and never treats its absence as a failure.

### Query text is never collected

Only the *duration* of the longest running query is read. The statement itself is
never queried, stored, or displayed — it is adjacent to personal data and belongs
to a different product category than uptime monitoring.

### Supported versions

PostgreSQL 12+ and MySQL 8.0+. On older servers the monitor works exactly as
before and the health figures are simply skipped — the page says so, so you can
tell "too old" from "missing grant". They ask for different fixes.

### On an incident

When a check fails and opens an incident, whatever health it had collected is
frozen onto that incident. The monitor page always shows the *current* figures;
the incident shows what they were *when it broke*, and those never change
afterwards even once the database recovers.
