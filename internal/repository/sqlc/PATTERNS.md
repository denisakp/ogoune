# sqlc Migration Patterns

Living document. Wave-1 (046) populated the worked examples below from the 7 migrated repos. Future-wave maintainers append entries as new patterns surface.

See also: [README.md](./README.md) (pilot walkthrough — 045 tags) and `specs/046-wave1-sqlc-crud/contracts/tx-and-helpers.md` for the call-site conventions.

---

## 1. Encryption call-sites (GORM hooks bypassed)

GORM hooks (`BeforeCreate/BeforeUpdate/AfterFind`) on encryption-bearing domain structs are **bypassed** by sqlc-generated code. The sqlc wrapper invokes `pkg/crypto` helpers at every write (before sending params) and every read (after scanning rows). Guard every call with `if len(field) > 0` — matches the GORM hook guards exactly.

**Canonical examples** (Wave 1):

- `internal/repository/store/notification_channel_repository_sqlc.go` — `encryptChannelConfig` / `decryptChannelConfig` wrap `crypto.EncryptChannelConfig` / `DecryptChannelConfig` on `Config []byte`. Applied in `Create`, `Update`, and all five read methods (`FindByID`, `List`, `FindByType`, `FindDefaultChannels`, `FindByResourceID`, `FindByComponentID`).
- `internal/repository/store/resource_credential_repository_sqlc.go` — two independent fields (`Password`, `Options`). Encrypt/decrypt each via dedicated helpers. Decryption failure returns sentinel `domain.ErrCredentialDecryption` (mirrors `domain/models.go:550-566`).

**Test pattern** (encryption round-trip — SC-006 gate):
- `internal/repository/store/notification_channel_repository_sqlc_test.go:TestNotificationChannelRepository_SqlcEncryption_RoundTrip` — write via wrapper → raw-read column → assert ciphertext → port read → assert plaintext.
- `internal/repository/store/resource_credential_repository_sqlc_test.go:TestResourceCredentialRepository_SqlcEncryption_RoundTrip` — same shape on `password` + `options` independently.

**APP_SECRET_KEY for tests**: set via `t.Setenv` + `crypto.SetGlobalProvider(&crypto.EnvKeyProvider{})`. Helpers: `setupChannelCryptoKey` / `setupCredentialCryptoForTest`.

---

## 2. SQL-native expressions (`CURRENT_TIMESTAMP`)

Wrappers MUST NOT compute timestamps in Go when the GORM impl used a SQL-native expression. Emit the same token literally in the query body; pass only non-timestamp params.

**Canonical example** (Wave 1):

- `internal/repository/sqlc/queries/postgres/user.sql` and `internal/repository/sqlc/queries/sqlite/user.sql` — `UpdateUserLastLogin :exec`:

  ```sql
  UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = $1
  ```

  Wrapper signature: `UpdateLastLogin(ctx, userID string) error` — passes only the ID, no `time.Now()`.

---

## 3. Singleton upserts (`ON CONFLICT (col) DO UPDATE`)

For singleton-config rows (one row keyed on a fixed col), use dialect-native upsert. Both Postgres and SQLite (≥3.24) support `ON CONFLICT(col) DO UPDATE`.

**Canonical examples** (Wave 1):

- **Read-or-create singleton** — `internal/repository/store/statuspage_settings_repository_sqlc.go`: `Upsert` reads existing via `GetStatusPageSettings`, then either calls `CreateStatusPageSettings` (if `pgx.ErrNoRows` / `sql.ErrNoRows`) or `UpdateStatusPageSettings`. The first inserted row's ID becomes "the singleton"; subsequent upserts preserve it. **Important**: `Get` on an empty table returns **default settings**, not `ErrNotFound` (mirrors GORM `domain/models.go` defaults).

- **`ON CONFLICT … DO UPDATE` for keyed upsert** — `internal/repository/sqlc/queries/postgres/resource_credential.sql` + sqlite mirror: `UpsertResourceCredential :exec`:

  ```sql
  INSERT INTO resource_credentials (...)
  VALUES ($1, $2, $3, $4, $5, $6, $7)
  ON CONFLICT (resource_id) DO UPDATE
  SET username = EXCLUDED.username,
      password = EXCLUDED.password,
      options = EXCLUDED.options,
      updated_at = EXCLUDED.updated_at;
  ```

  SQLite uses `excluded.col` (lowercase) instead of `EXCLUDED.col` for slightly safer parsing.

---

## 4. JSON columns (Go-side `json.Marshal` / `json.Unmarshal`)

When a column stores JSON but sqlc emits `string` or `[]byte`, the wrapper must marshal on write and unmarshal on read.

**Canonical example** (Wave 1):

- `internal/repository/store/incident_diagnostics_repository_sqlc.go` — `headersToJSON(map[string]string) (string, error)` and `headersFromJSON(string) (map[string]string, error)` for `request_headers` + `response_headers`. Empty/null handling: `nil` map → `"{}"`; empty string → `map[string]string{}`.

**Note**: `notification_channel.config` is already `[]byte` at the domain level (already JSON-encoded by callers). Wrapper forwards bytes through `crypto.*` directly without re-marshaling. JSON helper is only needed when the domain has a structured Go type ↔ JSON column.

---

## 5. Mapping helpers — type-shape reference

Per-type helpers live in the wrapper file that needs them. Cross-cutting helpers live in the file of the first wrapper that needed them.

| Field shape | Postgres (pgx) | SQLite (database/sql) | Domain |
|-------------|----------------|------------------------|--------|
| `time.Time` non-null | `pgtype.Timestamptz` (`.Time`) | `time.Time` | `time.Time` |
| `*time.Time` nullable | `pgtype.Timestamptz` (check `.Valid`) | `sql.NullTime` | `*time.Time` |
| `string` non-null | `string` | `string` | `string` |
| `*string` nullable | `pgtype.Text` (`.String` + `.Valid`) | `sql.NullString` | `*string` |
| `int` (PG `INTEGER`) | `int32` | `int64` (SQLite `INTEGER`) | `int` (cast in mapper) |
| `int64` | `int64` | `int64` | `int64` |
| `[]byte` (incl. encrypted blob) | `[]byte` | `[]byte` | `[]byte` |
| `bool` | `bool` | `int64` (SQLite has no bool) | `bool` (`!= 0` cast on SQLite) |
| `*bool` nullable (PG `BOOLEAN`) | `pgtype.Bool` | `sql.NullBool` | `*bool` |
| `*bool` nullable (PG `BOOLEAN` but SQLite `INTEGER` per migration) | `pgtype.Bool` | `sql.NullInt64` (cast `Int64 != 0`) | `*bool` |
| `*int` nullable (PG `INTEGER`) | `pgtype.Int4` (`.Int32`) | `sql.NullInt64` | `*int` |

**Cross-cutting helpers** (defined in the first wrapper that needed them; reused across Wave-1 files):

| Helper | File | Purpose |
|--------|------|---------|
| `pgTextFromPtr(*string) pgtype.Text` | `tags_repository_sqlc.go` (045) | `*string` → PG nullable text |
| `nullStringFromPtr(*string) sql.NullString` | `tags_repository_sqlc.go` (045) | `*string` → SQLite nullable text |
| `pgTimestampFromPtr(*time.Time) pgtype.Timestamptz` | `api_key_repository_sqlc.go` | `*time.Time` → PG nullable timestamp |
| `nullTimeFromPtr(*time.Time) sql.NullTime` | `api_key_repository_sqlc.go` | `*time.Time` → SQLite nullable timestamp |
| `boolToInt64(bool) int64` | `api_key_repository_sqlc.go` | `bool` → SQLite `INTEGER` (0/1) |
| `pgBoolFromPtr(*bool) pgtype.Bool` | `incident_diagnostics_repository_sqlc.go` | `*bool` → PG nullable bool |
| `pgInt4FromPtr(*int) pgtype.Int4` | `incident_diagnostics_repository_sqlc.go` | `*int` → PG nullable int |
| `nullBoolFromPtr(*bool) sql.NullBool` | `incident_diagnostics_repository_sqlc.go` | `*bool` → SQLite `BOOLEAN` (keyword_found case) |
| `nullBoolFromPtrAsInt64(*bool) sql.NullInt64` | `incident_diagnostics_repository_sqlc.go` | `*bool` → SQLite `INTEGER` (icmp_* case) |
| `nullIntFromPtr(*int) sql.NullInt64` | `incident_diagnostics_repository_sqlc.go` | `*int` → SQLite nullable int |

**Important migration drift caveat**: When the same logical type lands in SQLite as different column types across migrations (e.g., `icmp_available` was `INTEGER`, `keyword_found` was `BOOLEAN`), sqlc generates different Go types (`sql.NullInt64` vs `sql.NullBool`). The wrapper handles both. Future Wave-2/3 work should pick ONE convention per type (recommend `INTEGER` for SQLite bool to match sqlc's choice for the majority of pre-Wave-1 columns).

---

## 6. Behavior parity with GORM — when sqlc impl diverges

- **`gorm.Save()` upserts on missing rows**. Sqlc impl with `:execrows` returns `ErrNotFound` on zero-rows-affected. This is a divergence. Per FR-006 "when the port's contract requires it" — the port does NOT require ErrNotFound on Update for repos where the GORM impl upserts. Wave-1 chose: sqlc returns ErrNotFound (stricter, safer); contract tests do NOT assert this for `incident_diagnostics`, `notification_channel`, `user` (since GORM doesn't honor it).
- **`Update*` methods on `user_repository`** (`Update`, `Delete`, `UpdatePassword`, `UpdateLastLogin`, `UpdateTwoFactorSecret`) use `:exec` (NOT `:execrows`) because GORM's `Updates(map)` doesn't return ErrNotFound either. Match GORM exactly.
- **`statuspage_settings.Get` on empty table** returns DEFAULT settings (not ErrNotFound) — mirrors GORM impl precisely.
- **`expiry_notification_log.Delete*`** is bulk delete; no ErrNotFound on zero matches (GORM bulk delete behaves the same).

---

## 7. Bootstrap flag pattern

One env var per repo (`SQLC_<REPO>`). Selection helper shape: `selectXxxRepo(rt, db) (port.XxxRepository, string, error)`. Fail-fast when flag ON + dialect handle nil. Single shared `checkDialectHandle` helper.

**Canonical**: `internal/platform/bootstrap/database.go` — 8 selection helpers (1 from 045 + 7 from Wave 1), all identical shape.

---

## 8. CI lane

Wave 1 added one combined lane: `test-be-sqlc-wave1` (GitHub) / `backend-tests-sqlc-wave1` (GitLab). Sets all 7 env vars, runs `make test-be-pg` on both dialects via testcontainers. Failures localized by test name (`TestXxxRepository_SqlcContract`).

Future waves follow the same pattern: one combined lane per wave, not per repo.

---

## Anti-patterns

- **No cross-dialect tx interface.** Each `WithTx` helper takes its native handle. Don't introduce `TxRunner` interface (045 README).
- **No auto-generated mappers.** Manual mappers stay small and explicit.
- **No silent fallback** in bootstrap. Fail fast on flag-ON + nil handle.
- **Don't edit GORM impl** when shipping a sqlc wrapper. Both coexist until Wave 4.
- **Don't compute timestamps in Go** when the GORM impl uses SQL-native (`CURRENT_TIMESTAMP`, etc.).
- **VARCHAR(26) limit on resource_credentials.id** — IDs longer than 26 chars fail; if a column has a length constraint, the wrapper's caller must honor it (tests must use short IDs).

---

## 7. M2M write transactions (Wave 2)

True M2M wrappers (currently `maintenance_repository`) wrap multi-row writes in `pg.WithTx` / `sqlitesqlc.WithTx` (045 helpers). Pattern:

- **Create**: principal INSERT + N junction INSERTs inside one tx.
- **Update**: principal UPDATE + diff DELETE/INSERT on junction set, all in one tx.
- **Delete**: relies on FK `ON DELETE CASCADE` (verified in migration 0001). No Go-side junction DELETE.

Diff helper: `internal/repository/store/m2m_helpers.go:diffJunctionSets(current, target) (toAdd, toRemove)`.

**Canonical example**: `internal/repository/store/maintenance_repository_sqlc.go` — Create/Update wrap in `WithTx`; M2M rollback test at `maintenance_repository_sqlc_test.go:TestMaintenanceRepository_SqlcMM_TxRollback` (SC-008) injects FK violation, asserts principal row absent.

**Behavior divergence flagged**: GORM's `Save(m)` **appends** to M2M associations without removing diffs. The sqlc impl computes a proper diff. The contract test for `Update` only asserts principal-field semantics; the diff assertion is sqlc-specific (`TestMaintenanceRepository_SqlcMM_DiffOnUpdate`).

**Note**: `component_repository` originally appeared to be M2M (`component_notification_channels` junction), but the GORM impl preloads only `Resources` (one-to-many via `resources.component_id`). No junction writes from the component side. Sqlc impl is a plain CRUD wrapper.

---

## 8. ClaimPending dialect-divergent patterns (Wave 2)

`notification_repository.ClaimPending` is an atomic-claim hot path. Diverges per dialect:

- **Postgres**: `SELECT id … FOR UPDATE SKIP LOCKED` inside `pg.WithTx`, then `UPDATE … SET claim_owner, claimed_at`. SELECT either returns one locked row (we won) or zero (skipped). UPDATE under held lock is atomic.
- **SQLite**: single `UPDATE … WHERE id = ? AND status = 'pending' AND (claim_owner IS NULL OR claim_owner = '')`. SQLite's single-writer lock serializes; WHERE-guard ensures only one UPDATE matches.

**Test pattern** (SC-006 gate): `TestNotificationRepository_ClaimPending_ConcurrentSafety` spawns N=10 goroutines all calling `ClaimPending(ctx, sameID, …)`; asserts exactly-one-winner. Run 50× with `-race` for flake-flush.

**Canonical example**: `internal/repository/store/notification_repository_sqlc.go` (`ClaimPending` branches on driver).

**Validation contract**: Empty `id` or `claimOwner` → `repository.ErrInvalidInput`. Already-claimed or non-pending row → `(false, nil)`.

---

## 9. Dialect-divergent aggregations (Wave 2 chose Go-side)

For analytical queries (`monitoring_activity.GetUptimeStats` etc.), Wave 2 chose **Go-side aggregation over single SELECT** instead of dialect-divergent SQL (CTE/window vs. GROUP BY). Rationale:

- The GORM impl already does Go-side aggregation; sqlc impl mirroring it guarantees per-dialect numerical parity for free.
- Avoids `-- PG-only` / `-- SQLite-only` SQL drift maintenance.
- One Go code path → one parity gate (`TestMonitoringActivityRepository_Aggregations_GORMvsSQLC_SameDialect`, SC-007).

**Annotation convention** stays available for future repos that genuinely need dialect-divergent SQL (e.g., Wave-3 `resource_repository` may need PG-specific JSONB ops): `-- PG-only` / `-- SQLite-only` at the top of divergent queries.

**Canonical example**: `internal/repository/store/monitoring_activity_repository_sqlc.go` — all aggregations do Go-side bucketing/counting from base SELECTs. `GetAvgResponseTimeByWindow` uses native `AVG()` SQL but COUNT-first to avoid NULL scanning on empty windows.

**Raw-SQL migration**: `GetRecentResponseTimes` was raw SQL in the GORM impl; now a named `:many` sqlc query.

---

## 10. Preload via single JOIN (Wave 2)

For 1:1 PK-lookup preloads (e.g., `incident_event_step.FindByID` → embedded `Incident`), use a **single JOIN query** instead of two sequential queries. Clarification Q1 locked this — one round-trip beats two.

**SQL shape**:

```sql
SELECT
    s.id, s.created_at, s.updated_at, s.incident_id, s.step, s.message,
    i.created_at AS i_created_at,
    i.updated_at AS i_updated_at,
    i.resource_id, i.cause, i.resolved_at, i.started_at, i.details
FROM incident_event_steps s
JOIN incidents i ON i.id = s.incident_id
WHERE s.id = $1;
```

sqlc generates a flat row struct (`FindIncidentEventStepByIDRow`) with all step + incident columns (renamed via `AS x_alias` to avoid name collisions). Wrapper splits the flat row into `*domain.IncidentEventStep{Incident: domain.Incident{…}}`.

**SQLite parser caveat**: sqlc's sqlite parser fails on certain alias forms (numeric suffix like `_2`, complex AS aliases on the same name across tables). Workaround: rename aliases to underscored prefixes (`i_created_at` not `incident_id_2`).

**Canonical example**: `internal/repository/store/incident_event_step_repository_sqlc.go:eventStepFromPGJoin` / `eventStepFromSQLiteJoin`. Test: `TestIncidentEventStepRepository_SqlcJoinPreload`.

**For 1:many preloads** (maintenance.Resources): use **two queries** instead (principal + junction list). Wrapper assembles via stub resources (just IDs populated, full hydration deferred — service layer falls back to a separate `FindByComponentID` / `FindByID` lookup).
