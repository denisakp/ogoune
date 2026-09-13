# Technical debt register

Named so it can be prioritised. "Documented and not urgent" is not a state anyone else can
arbitrate; an entry here has a name, a consequence, and a recommendation. Urgency is not always a
technical judgement — that is the point of writing it down.

| # | Debt | Consequence | Recommendation | Since |
|---|---|---|---|---|
| 2 | `schema_migrations` is keyed on the numeric prefix alone (`0035`), not on the file name. | Two files sharing a number would be treated as one applied migration. Nothing in the drift check sees it. | **Do not fix.** Two guards already exist: the migrator refuses duplicate versions at startup, and `go test ./internal/database/` asserts it. Rewriting the state table to a name-keyed scheme is a migration of the migration system, on every installed database, for zero user-visible benefit. Recorded as a deliberate non-choice; revisit only if a third guard proves necessary. | 2026-09-10 |

Closed entries move to the bottom with the PR that closed them.

## Closed

| # | Debt | Closed by |
|---|---|---|
| 1 | `release.yml` ran three actions on Node 20 (`setup-go@v5`, `setup-qemu-action@v3`, `action-gh-release@v2`); the beta.6 release run already logged *"forced to run on Node.js 24"*. | Bumped to `setup-go@v6` (the major `ci.yml` already uses), `setup-qemu-action@v4`, `action-gh-release@v3` — each declares `node24`; every input the workflow passes to `action-gh-release` still exists in v3. Proven by the `v1.0.0-beta.7` release run. |

