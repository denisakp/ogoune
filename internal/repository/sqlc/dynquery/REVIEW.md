# Dynquery — SQL Injection Review Checklist

Any PR touching `internal/repository/sqlc/dynquery/` requires tech-lead review.
Use this checklist; if any item fails, reject the PR until fixed.

## Hard rules (any violation = block)

1. **No `fmt.Sprintf` / `+` building SQL strings.** User values reach the SQL
   only through `sq.Eq{...}`, `sq.NotEq{...}`, `sq.GtOrEq{...}`, `sq.LtOrEq{...}`,
   `sq.Like{...}`, or `sq.Expr("... ?", value)`. The placeholder `?` is the
   ONLY way a value enters SQL.
2. **Column names, JOIN clauses, ORDER BY columns, operators are hardcoded Go
   constants** — never derived from user input or filter struct fields.
3. **Validation precedes builder.** `MonitorFilter.Validate()` /
   `IncidentFilter.Validate()` must run on parsed input BEFORE
   `BuildXxxQuery`. Builder must never see invalid enum values.
4. **No new column names added without enum/whitelist check.** Adding `r.type`
   conditional → `isValidResourceType` must reject anything that is not in the
   hardcoded set.

## Soft rules (challenge in review)

5. **`sq.Or{...}` predicates carry parameterised values too.** A `LIKE` OR
   chain still uses `?` placeholders for the value.
6. **`LIKE` filters escape user input via `likeEscape` + `ESCAPE '\'`.** Bare
   `LIKE ?` would let users use `%` / `_` as wildcards (semantic correctness
   issue, not strictly injection).
7. **`PlaceholderFormat` is taken from the caller (sq.Dollar / sq.Question)**,
   not hardcoded. The builder is dialect-agnostic.

## Test gates (every PR)

- `go test ./internal/repository/sqlc/dynquery/...` includes:
  - SQL matrix tests (`monitors_test.go`, `incidents_test.go`) — assert
    payloads never appear inline as quoted literals in generated SQL.
  - Fuzz seed runs (`monitors_fuzz_test.go`, `incidents_fuzz_test.go`) — 20+
    SQLi/Unicode/null payloads per builder.
- `make fuzz-dynquery` — 30s × 2 campaigns, must pass before any release that
  touches this package.

## What this does NOT protect against

- Application-layer abuse (rate limiting — separate concern).
- Timing-side-channel attacks (not in scope).
- Authorised information disclosure (single-tenant by design).
- Bugs in the underlying driver. `pgx/v5` and `modernc.org/sqlite` are trusted
  components; if a driver-level injection ever lands, that's an upstream fix.

## Why a manual checklist instead of an automated source check?

A grep for `fmt.Sprintf` / `+` produces false positives (e.g. error messages,
comments) and false negatives (any sufficiently clever construction). The
3-layer defence (builder by construction → SQL matrix tests → fuzz tests) is
the actual safety net. This checklist anchors the review process, not a
brittle CI grep step.
