package strategy

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
)

// Database health collection against a REAL PostgreSQL server (spec 088).
//
// The existing protocol tests are unit tests over connection strings and error
// mapping; none of them speaks to an engine. Collection cannot be tested that
// way: the behaviour that matters is what a server actually returns to a role
// with and without a grant, and that is precisely what a mock would get wrong.
// These reuse the repository test fixture's container and skip without Docker.

type pgTarget struct {
	host     string
	port     int
	user     string
	password string
	database string
}

func parsePgDSN(t *testing.T, dsn string) pgTarget {
	t.Helper()
	u, err := url.Parse(dsn)
	require.NoError(t, err, "fixture DSN should be a URL")

	host := u.Hostname()
	port, err := strconv.Atoi(u.Port())
	require.NoError(t, err)
	pw, _ := u.User.Password()

	return pgTarget{
		host:     host,
		port:     port,
		user:     u.User.Username(),
		password: pw,
		database: strings.TrimPrefix(u.Path, "/"),
	}
}

// monitorFor builds a protocol monitor pointing at the fixture's own database.
func monitorFor(tgt pgTarget, user, password string) *domain.Resource {
	proto := "postgres"
	port := tgt.port
	return &domain.Resource{
		Target:       fmt.Sprintf("postgres://%s:%d/%s", tgt.host, tgt.port, tgt.database),
		Timeout:      10,
		ProtocolType: &proto,
		ProtocolPort: &port,
		Credential: &domain.ResourceCredential{
			Username: user,
			Password: []byte(password),
		},
	}
}

func setupPgTarget(t *testing.T) pgTarget {
	t.Helper()
	fx := internaltest.SetupPostgres(t)
	if fx == nil {
		t.Skip("postgres backend unavailable")
	}
	require.NotEmpty(t, fx.DSN, "fixture must expose a DSN")
	return parsePgDSN(t, fx.DSN)
}

// execAsSuperuser runs DDL on the fixture database as its owning role.
func execAsSuperuser(t *testing.T, tgt pgTarget, statements ...string) {
	t.Helper()
	ctx := context.Background()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		tgt.user, tgt.password, tgt.host, tgt.port, tgt.database)
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer conn.Close(ctx)

	for _, s := range statements {
		_, err := conn.Exec(ctx, s)
		require.NoErrorf(t, err, "statement: %s", s)
	}
}

// T016 -- the populated path, and the check is untouched by it.
func TestPostgresHealth_CollectsSaturation(t *testing.T) {
	tgt := setupPgTarget(t)
	r := monitorFor(tgt, tgt.user, tgt.password)

	res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, 10*time.Second, unsafeDialer)

	require.Equal(t, string(domain.StatusUp), res.Status, "collection must never change the verdict")
	assert.Empty(t, res.ErrorMessage)
	require.NotNil(t, res.DatabaseHealth)

	h := res.DatabaseHealth
	require.NotNil(t, h.ConnectionsActive, "available to any role that can connect")
	require.NotNil(t, h.ConnectionsMax, "the pair is present together or not at all")
	assert.Positive(t, *h.ConnectionsActive, "our own session is at least one")
	assert.Greater(t, *h.ConnectionsMax, *h.ConnectionsActive)
	assert.False(t, h.UnsupportedVersion, "the fixture runs a supported server")
	assert.Positive(t, h.CollectionDuration, "collection cost is measured, not assumed")
}

// T016 -- a monitor with no credential falls back to a TCP probe and collects
// nothing. Unchanged behaviour, asserted so it stays that way.
func TestPostgresHealth_NoCredentialCollectsNothing(t *testing.T) {
	tgt := setupPgTarget(t)
	proto := "postgres"
	port := tgt.port
	r := &domain.Resource{Target: tgt.host, Timeout: 5, ProtocolType: &proto, ProtocolPort: &port}

	res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, 5*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusUp), res.Status)
	assert.Nil(t, res.DatabaseHealth, "no session, nothing to collect")
}

// T019 -- a bad credential is an availability failure, exactly as before the
// feature. Health has nothing to do with it.
func TestPostgresHealth_BadCredentialStillFailsTheCheck(t *testing.T) {
	tgt := setupPgTarget(t)
	r := monitorFor(tgt, tgt.user, "definitely-not-the-password")

	res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, 10*time.Second, unsafeDialer)
	assert.Equal(t, string(domain.StatusDown), res.Status)
	assert.Nil(t, res.DatabaseHealth)
}

// T020 -- collection is abandoned at its allowance rather than allowed to finish.
// A zero budget is the degenerate case of the same rule and is what a check whose
// liveness ping already burned the deadline will see.
func TestPostgresHealth_ZeroBudgetSkipsCollection(t *testing.T) {
	tgt := setupPgTarget(t)
	ctx := context.Background()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		tgt.user, tgt.password, tgt.host, tgt.port, tgt.database)
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer conn.Close(ctx)

	assert.Nil(t, collectPostgresHealth(ctx, conn, 0), "no allowance, no attempt")
	assert.Nil(t, collectPostgresHealth(ctx, conn, -time.Second))
}

// T020 -- an allowance too small to complete the work returns without the fields
// and, crucially, returns *promptly*: the caller is never made to wait for work
// it has already given up on.
func TestPostgresHealth_ExhaustedContextDoesNotHang(t *testing.T) {
	tgt := setupPgTarget(t)
	ctx := context.Background()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		tgt.user, tgt.password, tgt.host, tgt.port, tgt.database)
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer conn.Close(ctx)

	start := time.Now()
	_ = collectPostgresHealth(ctx, conn, 30*time.Millisecond)
	assert.Less(t, time.Since(start), 2*time.Second,
		"collection returns on its own deadline, it does not run to completion")
}

// T041 -- THE test of this feature. A role that can connect but cannot read
// cluster-wide statistics must get the connection pair and NOTHING else.
//
// The failure this guards against is not a missing value, it is a plausible wrong
// one: without the grant, pg_stat_activity still returns other sessions' rows
// with their query_start withheld, so a naive maximum silently reports this
// monitor's own session age as the database's longest running query. In the right
// units, in a believable range, and undetectable by an operator.
func TestPostgresHealth_PartialVisibilityOmitsRatherThanUnderReports(t *testing.T) {
	tgt := setupPgTarget(t)
	execAsSuperuser(t, tgt,
		"DROP ROLE IF EXISTS ogoune_probe_limited",
		"CREATE ROLE ogoune_probe_limited LOGIN PASSWORD 'probe'",
		fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO ogoune_probe_limited", tgt.database),
	)
	t.Cleanup(func() {
		// The CONNECT grant is an object depending on the role, so it has to go
		// first or DROP ROLE fails.
		execAsSuperuser(t, tgt,
			fmt.Sprintf("REVOKE ALL ON DATABASE %s FROM ogoune_probe_limited", tgt.database),
			"DROP ROLE IF EXISTS ogoune_probe_limited",
		)
	})

	r := monitorFor(tgt, "ogoune_probe_limited", "probe")
	res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, 10*time.Second, unsafeDialer)

	require.Equal(t, string(domain.StatusUp), res.Status, "a missing grant costs a field, never the check")
	require.NotNil(t, res.DatabaseHealth)
	h := res.DatabaseHealth

	assert.NotNil(t, h.ConnectionsActive, "saturation needs no grant")
	assert.NotNil(t, h.ConnectionsMax)
	assert.Nil(t, h.LongestQuerySeconds,
		"omitted, not under-reported: a partial view would return our own session age and look right")
	assert.True(t, h.PrivilegeLimited, "so the interface can name the optional grant")
	assert.False(t, h.UnsupportedVersion, "a missing grant is not an old server; the operator needs different advice")
}

// T042 -- a database with no replication reports absence, never zero. Zero would
// read as "perfectly in sync", which is a different and much more reassuring
// claim than "not replicating".
func TestPostgresHealth_NoReplicationIsAbsentNotZero(t *testing.T) {
	tgt := setupPgTarget(t)
	r := monitorFor(tgt, tgt.user, tgt.password)

	res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, 10*time.Second, unsafeDialer)
	require.NotNil(t, res.DatabaseHealth)
	assert.Nil(t, res.DatabaseHealth.ReplicationLagSeconds,
		"the fixture has no standby; absent, never 0")
}

// T048 -- query text is never captured. The longest-running-query signal carries
// its duration and nothing else: statement text is PII-adjacent and collecting it
// would drag the product across the APM boundary the roadmap declares out of
// scope (FR-007).
func TestPostgresHealth_NeverCapturesQueryText(t *testing.T) {
	tgt := setupPgTarget(t)

	// A recognisable statement running in the background for the duration.
	const marker = "ogoune_secret_marker_value"
	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			tgt.user, tgt.password, tgt.host, tgt.port, tgt.database)
		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			return
		}
		defer conn.Close(context.Background())
		_, _ = conn.Exec(ctx, fmt.Sprintf("SELECT pg_sleep(2) /* %s */", marker))
	}()
	time.Sleep(300 * time.Millisecond)

	r := monitorFor(tgt, tgt.user, tgt.password)
	res := postgresCheck(context.Background(), r, tgt.host, tgt.port, false, 10*time.Second, unsafeDialer)
	<-done

	require.NotNil(t, res.DatabaseHealth)
	// The duration is collected...
	assert.NotNil(t, res.DatabaseHealth.LongestQuerySeconds,
		"a superuser sees the running statement's age")
	// ...and nothing anywhere in the result carries the statement itself.
	assert.NotContains(t, res.ResponseData, marker)
	assert.NotContains(t, res.ErrorMessage, marker)
	assert.NotContains(t, res.ResponseBody, marker)
	assert.NotContains(t, fmt.Sprintf("%+v", res.DatabaseHealth), marker,
		"no field of DatabaseHealth may hold statement text")
}
