# Compatibility

What Ogoune promises to keep stable, what it does not, and how a version number tells you which is which. This page is the contract the stable release will be held to; during the beta it already holds unless a release note says otherwise.

## Versions

Ogoune uses semantic versioning: `MAJOR.MINOR.PATCH`, with `-beta.N` pre-releases before 1.0.0.

- **PATCH** — fixes only. Nothing on this page changes.
- **MINOR** — additions. Everything covered below stays backward compatible: an integration written against 1.2 keeps working on 1.7.
- **MAJOR** — the only place a covered surface may change incompatibly, announced in the release notes with a migration path.
- **`-beta.N`** — the same rules, best effort. A beta may still change a covered surface; when it does, the release notes say so explicitly.

The server, the agent and the container images share one version number and are released together.

## Covered

### The v1 API — `/api/v1/`

- Endpoints, request fields and response fields are **additive** within a major: new endpoints and new fields may appear, existing ones keep their name, type and meaning.
- New response fields are optional or nullable. A client that ignores unknown fields is forward compatible.
- Error bodies follow RFC 9457 problem details; the `type` values are part of the contract.
- The authoritative description is the OpenAPI document the server serves at `GET /api/v1/openapi.json`. The [API reference](/api/reference) is rendered from it.

### The agent wire protocol

The agent streams frames over a WebSocket at `/api/v1/agent/stream`. Each frame carries a `schema_version`.

- A server accepts every schema version **up to** the one it knows. A newer server always accepts an older agent.
- Additions ride on the existing frame as optional keys, which an older server ignores. A newer agent therefore keeps streaming to an older server; what it adds is simply not stored until the server is upgraded.
- The schema version moves only when a frame could no longer be read safely by a server that does not know the change. It has moved once (1 → 2, kernel events), and adding the capability declaration in beta.7 deliberately did not move it.

You can upgrade the server and the agent in either order. The [agent page](/self-host/agent) says what each side shows in the meantime.

### Agent configuration

The agent's flags (`-backend-url`, `-credential`, `-interval`, `-log-level`, `-insecure`, `-config`), the matching `OGOUNE_*` environment variables, and the `/etc/ogoune/agent.cfg` file format. Precedence — file, then environment, then flags — is part of the contract.

### Server configuration

The environment variables documented on the [configuration page](/self-host/configuration). A variable may gain a value or a sibling; it does not change meaning or disappear within a major. A variable that is going away is logged as deprecated for at least one minor version first.

### The database, as an upgrade path

- Migrations run automatically at startup, **forward only**, and are additive: they add tables and columns and never rewrite or delete existing rows. Every build is tested by upgrading a real database written by an earlier release — see [Upgrading](/self-host/upgrading).
- There is no downgrade path. Rolling back a version means restoring the backup you took before upgrading.
- The server refuses to start, with a message naming the files, rather than run an ambiguous migration set.

### Supported platforms

- **Server**: Linux `amd64` and `arm64` containers; SQLite (bundled) or PostgreSQL 16 and later; Redis 7 for the production scheduler.
- **Agent**: Linux `amd64` and `arm64`, as a container or a static binary. macOS and Windows builds run for development only and declare every kernel capability unavailable.

## Not covered

- **`/api/`** (non-versioned) — internal to the web interface. It may change in any release.
- **The database schema itself.** The upgrade path is covered; querying the tables directly is not. Column names, types and internal state tables (`schema_migrations`, `incident_diagnostics`) may change between minor versions.
- **Go packages** under `internal/` and `pkg/` — not a library API.
- **Log line formats** and Prometheus metric names, until a release note declares a metric stable.
- **Kernel-event capture coverage.** The agent declares what it can observe on a given machine; which signals a given kernel or container runtime exposes is not something Ogoune can promise. See [the agent page](/self-host/agent#kernel-events).

## What a beta can still change

Before 1.0.0 the surfaces above are treated as covered, and to date no beta has broken one. If a beta must, the release notes will say **Breaking** in the first line of the entry, with what to change. Two things in the beta series are permanent by design rather than breaking:

- Incidents created before v1.0.0-beta.7 have no recorded machine and no recorded agent capabilities. They show the monitor's current machine, marked *inferred*, and *not known* for capabilities. This cannot be backfilled; nothing was recorded to backfill from.
- Hosts running an agent older than v1.0.0-beta.7 show *this agent version does not report what it can observe* until the agent is upgraded.
