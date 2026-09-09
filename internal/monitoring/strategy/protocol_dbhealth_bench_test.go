package strategy

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

// What health collection costs a database check (spec 088, SC-003), measured
// against a real PostgreSQL rather than estimated.
//
// The two benchmarks differ only in whether collection runs, so their delta is
// the enrichment and nothing else.

const dbHealthBenchIterations = 200

func benchDBCheck(b *testing.B, name string, collect bool) {
	tgt := setupPgTargetB(b)
	r := monitorFor(tgt, tgt.user, tgt.password)
	// A zero timeout yields a zero allowance, which is how "no collection" is
	// expressed -- the same path a check with nothing left on its deadline takes.
	timeout := 10 * time.Second
	if !collect {
		r.Timeout = 0
	}

	noop := func(dbHealthSkipReason) {}
	for i := 0; i < 20; i++ {
		postgresCheck(context.Background(), r, tgt.host, tgt.port, false, timeout, unsafeDialer, noop)
	}

	durations := make([]time.Duration, 0, dbHealthBenchIterations)
	b.ResetTimer()
	for i := 0; i < dbHealthBenchIterations; i++ {
		start := time.Now()
		res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, timeout, unsafeDialer, noop)
		durations = append(durations, time.Since(start))
		if res.Status != string(domain.StatusUp) {
			b.Fatalf("iter %d: check failed, which collection must never cause", i)
		}
	}
	b.StopTimer()

	reportP95DB(b, name, durations)
}

func BenchmarkDatabaseCheck_WithoutHealth_p95(b *testing.B) {
	benchDBCheck(b, "BenchmarkDatabaseCheck_WithoutHealth", false)
}

func BenchmarkDatabaseCheck_WithHealth_p95(b *testing.B) {
	benchDBCheck(b, "BenchmarkDatabaseCheck_WithHealth", true)
}

// SC-003a, the one that matters most: on a database slow enough that collection
// times out every single time, 100% of checks still pass.
//
// This is the failure mode that would make the whole feature actively harmful --
// a *slow* database reported as a *down* database, by the very code meant to
// explain the slowness. It is asserted rather than assumed.
func TestDatabaseCheck_CollectionTimeoutNeverFailsTheCheck(t *testing.T) {
	tgt := setupPgTarget(t)
	r := monitorFor(tgt, tgt.user, tgt.password)
	// A timeout small enough that the allowance lands under the floor, so
	// collection is abandoned every time.
	r.Timeout = 1

	noop := func(dbHealthSkipReason) {}
	const runs = 25
	passed := 0
	for i := 0; i < runs; i++ {
		res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, 40*time.Millisecond, unsafeDialer, noop)
		if res.Status == string(domain.StatusUp) {
			passed++
		}
	}

	assert.Equal(t, runs, passed,
		"a database too slow to introspect is still a database that is up")
}

// setupPgTargetB is setupPgTarget for benchmarks. internaltest.SetupPostgres
// already takes a testing.TB, so only the DSN parsing differs.
func setupPgTargetB(b *testing.B) pgTarget {
	b.Helper()
	fx := internaltest.SetupPostgres(b)
	if fx == nil || fx.DSN == "" {
		b.Skip("postgres backend unavailable")
		return pgTarget{}
	}
	return parsePgDSNRaw(b, fx.DSN)
}

func reportP95DB(b *testing.B, name string, durations []time.Duration) {
	b.Helper()
	if len(durations) == 0 {
		return
	}
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	p := func(q float64) int64 {
		idx := int(float64(len(sorted)-1) * q)
		return sorted[idx].Microseconds()
	}
	line := fmt.Sprintf("db_health_bench name=%s iterations=%d p50_us=%d p95_us=%d p99_us=%d",
		name, len(sorted), p(0.50), p(0.95), p(0.99))
	fmt.Println(line)
	b.Log(line)
}

func parsePgDSNRaw(tb testing.TB, dsn string) pgTarget {
	tb.Helper()
	conn, err := pgx.ParseConfig(dsn)
	require.NoError(tb, err)
	return pgTarget{
		host:     conn.Host,
		port:     int(conn.Port),
		user:     conn.User,
		password: conn.Password,
		database: conn.Database,
	}
}
