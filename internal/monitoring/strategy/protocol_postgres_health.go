package strategy

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/denisakp/ogoune/internal/domain"
)

// PostgreSQL health collection (spec 088). Runs on the connection the check has
// already opened, after the liveness ping has succeeded, under its own allowance.
//
// The contract, in one line: this may cost a field, never a check. Every failure
// path here returns what it has and leaves the caller's result untouched.

// pgMinSupportedVersion is the oldest server this enrichment reads. Below it the
// fields are omitted and unsupported_version is set -- not a failure, and not a
// misconfiguration. 12 is the oldest branch still supported upstream; supporting
// older ones would mean conditional query shapes for servers nobody should expose
// to production monitoring.
const pgMinSupportedVersion = 120000

// collectPostgresHealth gathers what the server says about itself. Returns nil
// when nothing at all could be read.
func collectPostgresHealth(ctx context.Context, conn *pgx.Conn, budget time.Duration) *domain.DatabaseHealth {
	if budget <= 0 {
		return nil
	}
	start := time.Now()

	hctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	h := &domain.DatabaseHealth{}

	// Version first: below the floor there is no point issuing anything else, and
	// "too old" is a different message to the operator than "missing grant".
	var serverVersion int
	if err := conn.QueryRow(hctx, "SELECT current_setting('server_version_num')::int").Scan(&serverVersion); err != nil {
		return nil
	}
	if serverVersion < pgMinSupportedVersion {
		h.UnsupportedVersion = true
		h.CollectionDuration = time.Since(start)
		return h
	}

	// Connection saturation. Available to any role that can connect, which is why
	// User Story 1 ships without asking for a grant. Both values or neither: a
	// saturation ratio needs the pair.
	var active, max int64
	err := conn.QueryRow(hctx, `
		SELECT (SELECT count(*) FROM pg_stat_activity),
		       current_setting('max_connections')::bigint`).Scan(&active, &max)
	if err == nil {
		h.ConnectionsActive = &active
		h.ConnectionsMax = &max
	}

	// Everything past here is privilege-gated and belongs to User Story 3.
	collectPostgresPrivilegedHealth(hctx, conn, h)

	h.CollectionDuration = time.Since(start)
	if h.ConnectionsActive == nil && h.LongestQuerySeconds == nil && h.ReplicationLagSeconds == nil && !h.UnsupportedVersion {
		// Nothing readable at all -- report absence rather than an empty shell.
		return nil
	}
	return h
}

// collectPostgresPrivilegedHealth adds the two signals that need a grant. Written
// as a separate step because its failure mode is the dangerous one in this whole
// feature, and it deserves to be read on its own.
//
// A role without pg_read_all_stats still sees other sessions' ROWS in
// pg_stat_activity -- the columns it needs are withheld, not the rows. A naive
// max(now() - query_start) therefore returns this monitor's OWN session age: a
// number in the right units, in a plausible range, and completely wrong. The
// operator has no way to tell. So visibility is established first, and the field
// is omitted rather than computed over a partial view (FR-013).
func collectPostgresPrivilegedHealth(ctx context.Context, conn *pgx.Conn, h *domain.DatabaseHealth) {
	var canReadAllStats bool
	if err := conn.QueryRow(ctx,
		"SELECT pg_has_role(current_user, 'pg_read_all_stats', 'member') OR current_setting('is_superuser') = 'on'").
		Scan(&canReadAllStats); err != nil {
		h.PrivilegeLimited = true
		return
	}
	if !canReadAllStats {
		h.PrivilegeLimited = true
		return
	}

	// Longest running query: its AGE only. The statement text is never read,
	// never stored and never returned -- it is PII-adjacent and it would drag the
	// product across the APM boundary the roadmap declares out of scope (FR-007).
	var longest *float64
	if err := conn.QueryRow(ctx, `
		SELECT EXTRACT(EPOCH FROM max(now() - query_start))::float8
		FROM pg_stat_activity
		WHERE state = 'active' AND query_start IS NOT NULL AND pid <> pg_backend_pid()`).
		Scan(&longest); err == nil && longest != nil {
		h.LongestQuerySeconds = longest
	}

	collectPostgresReplicationLag(ctx, conn, h)
}

// collectPostgresReplicationLag sources lag from whichever side of replication
// this instance is on. pg_stat_replication lists clients and is populated on a
// PRIMARY; it is empty on a standby, so a monitor pointed at a replica that only
// queried it would report "no replication lag" for a replica hours behind. The
// standby's own lag comes from its replay timestamp instead.
//
// Absent when the instance is not replicating at all -- never 0, which would mean
// "perfectly in sync" (FR-015).
func collectPostgresReplicationLag(ctx context.Context, conn *pgx.Conn, h *domain.DatabaseHealth) {
	var inRecovery bool
	if err := conn.QueryRow(ctx, "SELECT pg_is_in_recovery()").Scan(&inRecovery); err != nil {
		return
	}

	var lag *float64
	if inRecovery {
		// Standby: how far behind its own replay is. Null while it has replayed
		// everything it has received, which is not the same as "not replicating".
		err := conn.QueryRow(ctx, `
			SELECT CASE WHEN pg_last_wal_receive_lsn() = pg_last_wal_replay_lsn() THEN 0
			            ELSE EXTRACT(EPOCH FROM now() - pg_last_xact_replay_timestamp())::float8
			       END`).Scan(&lag)
		if err != nil {
			return
		}
	} else {
		// Primary: the worst lag across connected standbys. No rows means no
		// standbys, which is the overwhelmingly common case and must stay absent.
		err := conn.QueryRow(ctx, `
			SELECT EXTRACT(EPOCH FROM max(now() - reply_time))::float8
			FROM pg_stat_replication`).Scan(&lag)
		if err != nil {
			return
		}
	}
	if lag != nil {
		h.ReplicationLagSeconds = lag
	}
}
