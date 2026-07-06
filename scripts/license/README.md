# License guards

Three guards enforce the open-core licensing contract on every PR targeting `main`:

| Guard | Script | Implements | Contract |
|---|---|---|---|
| SPDX coverage | `check-spdx.sh` | FR-004, FR-005, SC-002 | `specs/040-relicense-apache-ee/contracts/spdx-coverage.contract.md` |
| Runtime-deps license | `check-deps.sh` | FR-008, SC-003 | `specs/040-relicense-apache-ee/contracts/deps-license.contract.md` |
| Documentation drift | `check-docs.sh` | FR-007, SC-004 | `specs/040-relicense-apache-ee/contracts/docs-drift.contract.md` |

Run all three locally:

```bash
make license-audit
```

Or run any guard individually:

```bash
scripts/license/check-spdx.sh
scripts/license/check-deps.sh
scripts/license/check-docs.sh
```

## Required tools

| Tool | Used by | Install |
|---|---|---|
| `go-licenses` | `check-deps.sh` | `go install github.com/google/go-licenses@v1.6.0` |
| `pnpm` | `check-deps.sh` | already used in `web/` (see `web/.npmrc`) |
| `jq` | `check-deps.sh` | `brew install jq` / `apt install jq` |
| `grep` | `check-spdx.sh`, `check-docs.sh` | system |

`go-licenses` is pinned at v1.6.0 in CI; the local copy should match. If a newer version changes flag names, update both this README and `.github/workflows/license-guards.yml` in the same PR.

## Configuration files

- `allowed-deps-licenses.txt` — SPDX identifiers explicitly accepted on the core scope.
- `denied-deps-licenses.txt` — SPDX identifier families that fail the build.
- `docs-allowlist.txt` — phrases that exempt a line from the docs-drift scan (used for historical / explanatory AGPL mentions).

Each list is plain text, one entry per line, `#`-prefixed lines treated as comments. Update with care; every addition is a deliberate policy change.
