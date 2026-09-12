# Technical debt register

Named so it can be prioritised. "Documented and not urgent" is not a state anyone else can
arbitrate; an entry here has a name, a consequence, and a recommendation. Urgency is not always a
technical judgement — that is the point of writing it down.

| # | Debt | Consequence | Recommendation | Since |
|---|---|---|---|---|
| 1 | Three GitHub Actions still run on Node 20: `actions/setup-go@v5`, `docker/setup-qemu-action@v3`, `softprops/action-gh-release@v2`. | The warning is already in the logs: the v1.0.0-beta.6 release run (2026-09-10) reports *"actions target Node.js 20 but are being forced to run on Node.js 24"*. It works because GitHub forces it; the day the forcing stops, the release workflow stops publishing — a delivery risk, not a code risk. `ci.yml` already uses `setup-go@v6`; all three old majors sit in `release.yml`. | Bump the three majors in `.github/workflows/release.yml` (one line each) and prove it with a tagged pre-release before the next real one. Cheap; the only reason it is open is that nobody has cut a release since the warning appeared. | 2026-09-10 |
| 2 | `schema_migrations` is keyed on the numeric prefix alone (`0035`), not on the file name. | Two files sharing a number would be treated as one applied migration. Nothing in the drift check sees it. | **Do not fix.** Two guards already exist: the migrator refuses duplicate versions at startup, and `go test ./internal/database/` asserts it. Rewriting the state table to a name-keyed scheme is a migration of the migration system, on every installed database, for zero user-visible benefit. Recorded as a deliberate non-choice; revisit only if a third guard proves necessary. | 2026-09-10 |

Closed entries move to the bottom with the PR that closed them.
