# Introduction

Ogoune is an uptime monitoring app that **confirms failures before alerting**. Instead of paging you on the first failed check, it requires _N consecutive failures_ before opening an incident — so a 2-second network blip never wakes you up.

## Open-core model

| | Community Edition | Enterprise Edition |
|---|---|---|
| License | Apache 2.0 | LicenseRef-Ogoune-EE |
| Database | SQLite (in-process) | PostgreSQL |
| Scheduler | TimingWheel (in-process) | Asynq (Redis) |
| Scaling | Single binary | Stateless API + external workers |

Both editions share the same codebase. Enterprise features are documented here in the open.

## Where to go next

- [Quickstart](/guide/quickstart) — run Ogoune locally in minutes
- [Core concepts](/guide/concepts) — monitors, checks, incidents
- [Self-hosting](/self-host/) — deploy for real
- [Cloud](/cloud/) — let us host it for you
