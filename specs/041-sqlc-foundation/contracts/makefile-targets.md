# Contract — Makefile Targets

## New / modified targets

### `SQLC_VERSION` (variable)

```make
SQLC_VERSION := v1.27.0
```

Single source of truth for sqlc CLI version. Bumping this variable is the only way to upgrade.

### `sqlc-bin` (internal helper)

**Behavior**: Ensures the pinned `sqlc` binary is on `PATH`. Installs via `go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)` if missing.

**Exit codes**: 0 on success; non-zero if install fails.

### `sqlc-generate`

**Behavior**: `sqlc-bin` → run `sqlc generate` per `sqlc.yaml`. Writes generated Go files under `internal/repository/sqlc/pg/` and `internal/repository/sqlc/sqlite/`.

**Exit codes**: 0 on successful generation; non-zero on sqlc errors.

**Side effects**: Modifies committed generated files. Maintainer commits the diff.

### `sqlc-check`

**Behavior**: `sqlc-bin` → `sqlc generate` → `git diff --exit-code -- internal/repository/sqlc/pg internal/repository/sqlc/sqlite`.

**Exit codes**: 0 if no drift; non-zero with a clear message (`run 'make sqlc-generate'`) if drift detected.

**Side effects**: Locally regenerates files; CI runners are ephemeral so drift surfaces immediately. On developer machines, drift left in the working tree is a feature (signals "you forgot to commit").

### `build-be` (modified)

**Behavior**: Depends on `sqlc-check` (pre-step). Fails fast before compilation if generated code is out of sync.

### `test-be`, `lint`, other targets

**Behavior**: Unchanged.

## CI integration

The existing CI workflow gains one step (executed once, dialect-agnostic):

```yaml
- name: sqlc check
  run: make sqlc-check
```

Placed before the test matrix so drift fails the pipeline early.
