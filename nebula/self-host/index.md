# Self-hosting

Ogoune runs in two modes depending on your scale.

| | Community | Production |
|---|---|---|
| Database | SQLite (in-process) | PostgreSQL |
| Scheduler | TimingWheel (in-process) | Asynq (Redis) |
| Scaling | Single binary | Stateless API + external workers |

- [Community (SQLite)](/self-host/community) — one binary, zero external deps
- [Production (Postgres + Redis)](/self-host/production) — horizontal scale
- [Configuration](/self-host/configuration) — environment variables
- [Upgrading](/self-host/upgrading) — backup, order, what is guaranteed, per-version notes
- [Compatibility](/self-host/compatibility) — what semver covers, and what it does not
