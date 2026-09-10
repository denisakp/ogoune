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

// SC-003a, the one that matters most: collection being abandoned must never
// turn an UP check into a DOWN one.
//
// This is the failure mode that would make the whole feature actively harmful --
// a *slow* database reported as a *down* database, by the very code meant to
// explain the slowness.
//
// The deadline here does double duty, and that is what the assertions have to
// account for: postgresCheck takes ONE timeout, which bounds both the connection
// and, at a quarter of its length, the health allowance. 40ms puts the allowance
// under the 20ms floor -- which is the point -- but it also gives the TCP connect
// and handshake 40ms, and on a loaded machine that is sometimes not enough.
//
// So this asserted 25 successes out of 25 and failed in CI for a reason that had
// nothing to do with health collection. Requiring every connection to a
// containerised database to complete within 40ms is not a property of this
// feature, and a test that fails on the machine rather than on the code teaches
// people to ignore it.
//
// What IS asserted, on every run and independent of load:
//
//   - a check that connected is UP and carries no health, which is the evidence
//     collection was abandoned rather than merely fast;
//   - a check that failed did so because of the CONNECTION. Nothing else may
//     report down at this deadline -- and structurally nothing can, since the
//     result is already UP before collection is even attempted, which is exactly
//     the guarantee under test.
func TestDatabaseCheck_CollectionTimeoutNeverFailsTheCheck(t *testing.T) {
	tgt := setupPgTarget(t)
	r := monitorFor(tgt, tgt.user, tgt.password)
	r.Timeout = 1

	noop := func(dbHealthSkipReason) {}
	const runs = 25
	connected, tooSlowToConnect := 0, 0

	for i := 0; i < runs; i++ {
		// A timeout small enough that the allowance lands under the floor, so
		// collection is abandoned every time.
		res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, 40*time.Millisecond, unsafeDialer, noop)

		if res.Status == string(domain.StatusUp) {
			connected++
			assert.Nilf(t, res.DatabaseHealth,
				"run %d: the allowance was under the floor, so collection must have been abandoned", i)
			continue
		}

		// Down. At this deadline the only honest reason is that the connection
		// itself did not fit in it -- the machine, not the feature.
		require.NotNilf(t, res.Cause, "run %d: a failed check must say why: %s", i, res.ResponseData)
		switch *res.Cause {
		case domain.ConnectionTimeout, domain.ProtocolHandshakeFailed:
			tooSlowToConnect++
		default:
			t.Fatalf("run %d: check reported down with cause %q (%s); "+
				"health collection must never be able to fail a check",
				i, *res.Cause, res.ResponseData)
		}
	}

	t.Logf("connected in %d of %d runs; %d could not connect within the 40ms deadline",
		connected, runs, tooSlowToConnect)

	// Without a single connection the run proves nothing, and a test that can
	// pass vacuously is worse than one that fails.
	require.NotZerof(t, connected,
		"not one of %d checks connected to PostgreSQL within 40ms; "+
			"this machine is too slow to exercise the property, not evidence against it", runs)
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
