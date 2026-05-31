package internaltest

import (
	"context"
	"testing"

	"github.com/denisakp/ogoune/internal/database"
)

// SetupPostgres returns a per-test Postgres fixture or, when no Postgres
// backend is available, calls t.Skip and returns nil.
//
// Availability sources (first-match wins):
//  1. POSTGRES_TEST_DSN env var — used as a superuser DSN; no Docker boot.
//  2. testcontainers-go (postgres:16-alpine) — requires Docker.
//
// Per-test isolation: the package-wide container provisions a fresh database
// cloned from a migrated template; t.Cleanup drops it.
func SetupPostgres(t *testing.T) *DialectFixture {
	t.Helper()
	c, skipReason := getPgContainer(t)
	if c == nil {
		t.Skipf("postgres backend unavailable: %s (set POSTGRES_TEST_DSN or run Docker)", skipReason)
		return nil
	}
	dsn := c.Acquire(t)

	rt, err := database.Open(context.Background(), database.Config{
		Driver:      database.DriverPostgres,
		DatabaseURL: dsn,
		LogLevel:    "silent",
	})
	if err != nil {
		t.Fatalf("internaltest: open postgres fixture: %v", err)
	}
	t.Cleanup(func() {
		if rt == nil {
			return
		}
		if sqlDB, dbErr := rt.GormDB().DB(); dbErr == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
		if rt.PgxPool() != nil {
			rt.PgxPool().Close()
		}
	})

	return &DialectFixture{Dialect: "postgres", Runtime: rt, DSN: dsn}
}
