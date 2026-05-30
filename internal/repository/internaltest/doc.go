// Package internaltest provides dual-dialect test helpers for repository
// tests. It exposes SetupSQLite, SetupPostgres, and ForEachDialect so a
// single test body can validate any repository implementation against both
// SQLite and Postgres backends with per-test database isolation.
//
// This package is test-only. It MUST NOT be imported from production code
// under cmd/, internal/api/, internal/service/, or non-_test.go files in
// internal/repository/store/. The CI polish phase asserts:
//
//	go list -deps ./cmd/... | grep -E 'testcontainers|internaltest'   # returns empty
//
// See specs/044-test-infrastructure/ for the design.
package internaltest
