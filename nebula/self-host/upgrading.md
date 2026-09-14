# Upgrading

Upgrades are designed to be boring: pull the new version, start it, done. This page is what "done" rests on, and the few things worth knowing per version.

## Before you start

**Take a backup.** Migrations are additive and tested, and there is still no downgrade path — a rollback is a restore.

```bash
# SQLite: stop the server first, then copy the file (and its -wal / -shm if present)
cp ogoune.db ogoune.db.bak

# PostgreSQL
pg_dump "$DATABASE_URL" > ogoune-$(date +%F).sql
```

## The order

1. **Server first.** Start the new version. It runs the pending migrations at startup and refuses to serve traffic if any fails — the log names the migration and what to do. A server that says `API server listening` has finished migrating.
2. **Agent second**, on every host. Not urgent: a newer server accepts an older agent unchanged. Until the agent is upgraded, its host page shows *this agent version does not report what it can observe*, and features that live in the agent (filesystem grouping, kernel capabilities) are absent for that host.

The reverse order also works — a newer agent streams to an older server — but what the new agent sends beyond what the old server knows is not stored.

## What is guaranteed

Every build is tested by restoring a database **written by an earlier released version** (both SQLite and PostgreSQL), running the new migrations against it, and checking that every table survives with the same rows, byte-for-byte on the columns that existed before. The oldest release covered is v1.0.0-beta.4. A migration that rewrote or dropped existing data would not pass CI.

Migrations are forward-only and additive. A migration file is never edited after release; a correction is a new migration.

## Rolling back

Stop the new version, restore the backup, start the previous version. The previous version does not know the new columns and does not need to — it never reads them — but the safe path is the restore, not running an older binary on a newer schema.

## Per-version notes

Only what changes for an operator. Full detail is in the [release notes](https://github.com/denisakp/ogoune/releases).

### v1.0.0-beta.7
- Two migrations, both additive: an incident records the machine and the agent capabilities it saw when it opened. Nothing to configure.
- Incidents from before this version show *inferred* for the machine and *not known* for capabilities, permanently — there is nothing recorded to backfill from.
- Upgrade the agent to get the capability declaration and one-entry-per-filesystem host metrics.

### v1.0.0-beta.6
- One migration, a renumbered file (`0026_report_settings` → `0034`) that re-applies as a no-op on databases that already ran it.
- The server now refuses to start when two migration files share a version number. A fresh install is unaffected.

### v1.0.0-beta.5
- **Security.** Fixes an unauthenticated account takeover through the password-initialisation endpoint (GHSA-9jv8-xmw9-c9f3, critical). Every install on beta.4 or earlier should upgrade immediately.
- There is no default admin password any more. On first boot without `AUTH_PASSWORD`, the server generates one and prints it in the log; set `AUTH_PASSWORD` to make it permanent.
- If you run the host agent, upgrade both server and agent.

### Earlier betas
- Upgrade to beta.5 or later before anything else, for the security fix above.
