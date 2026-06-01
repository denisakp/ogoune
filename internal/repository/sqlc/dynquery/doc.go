// Package dynquery builds dynamic, parameterised SQL for v1 list endpoints
// that accept query-param filters.
//
// Design (see specs/051-dynamic-filters/contracts/dynquery-package.md):
//
//   - One filter struct + one Build function per endpoint.
//   - Build functions take a sq.PlaceholderFormat so the same builder works
//     for Postgres (sq.Dollar) and SQLite (sq.Question).
//   - Column names, JOINs, and operators are HARDCODED CONSTANTS in the build
//     functions. User-derived values reach SQL exclusively through squirrel's
//     parameterised helpers (sq.Eq{}, sq.Like{}, sq.Expr("... ?", v)).
//   - Filter structs use pointer fields so a nil pointer == "no filter on
//     this field"; the absence of a query param maps to a nil pointer.
//
// Three-layer SQL-injection defence (per contracts/injection-safety.md):
//
//  1. Builder-by-construction (this package).
//  2. SQL-capture tests (monitors_test.go / incidents_test.go) that assert
//     payloads never appear inline in the generated SQL.
//  3. Fuzz tests (monitors_fuzz_test.go / incidents_fuzz_test.go).
package dynquery
