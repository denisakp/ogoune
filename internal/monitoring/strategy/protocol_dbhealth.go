package strategy

import (
	"context"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// Database health collection (spec 088) runs on the session a PostgreSQL or MySQL
// protocol check has already opened, between the liveness ping succeeding and the
// check returning.
//
// It gets its own time allowance rather than sharing the check's deadline. The
// work item said to share it, and sharing the *connection* is right — but both
// engine checks run connect and ping under one context, so an introspection query
// on a saturated database (the exact condition this feature exists to reveal)
// could push the whole check past its deadline and mark a slow database as down.
// A fraction of the operator's own timeout rather than a fixed ceiling, so a
// monitor deliberately given a long timeout is not starved and a short one cannot
// be overrun.
const (
	// dbHealthBudgetFraction is the share of the monitor's configured timeout that
	// collection may use.
	dbHealthBudgetFraction = 4 // i.e. 25%
	// dbHealthBudgetCap bounds the allowance in absolute terms, so a monitor with
	// a very generous timeout still cannot spend seconds on introspection.
	dbHealthBudgetCap = 2 * time.Second
	// dbHealthBudgetFloor is the point below which collection is not worth
	// attempting: the round trip alone would consume it.
	dbHealthBudgetFloor = 20 * time.Millisecond
)

// dbHealthBudget returns how long collection may run, given the check's parent
// context and the monitor's configured timeout. It is the smallest of: a quarter
// of that timeout, the absolute cap, and the time actually left on the context.
//
// Returns 0 when there is not enough left to be worth trying — the caller then
// skips collection entirely and the check proceeds untouched.
//
// Not configurable, deliberately: it derives from the timeout the operator has
// already set, so this feature adds nothing for them to tune.
func dbHealthBudget(ctx context.Context, timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 0
	}

	budget := timeout / dbHealthBudgetFraction
	if budget > dbHealthBudgetCap {
		budget = dbHealthBudgetCap
	}

	// Never promise more than the context actually has left. The liveness check
	// has already spent part of it, and on a slow database it may have spent most.
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < budget {
			budget = remaining
		}
	}

	if budget < dbHealthBudgetFloor {
		return 0
	}
	return budget
}

// dbHealthSkipReason names why collection produced nothing, for the metrics
// counter. Absence is invisible by design, so an operator needs to be able to see
// how often it happens and why without reading logs.
type dbHealthSkipReason string

const (
	dbHealthSkipDeadline    dbHealthSkipReason = "deadline"
	dbHealthSkipPrivilege   dbHealthSkipReason = "privilege"
	dbHealthSkipUnsupported dbHealthSkipReason = "unsupported_version"
	dbHealthSkipNoFields    dbHealthSkipReason = "no_fields"
)

// dbHealthSkipFunc reports one skip reason. Passed rather than reached for, so
// the engine check files stay free of any dependency on metrics.
type dbHealthSkipFunc func(dbHealthSkipReason)

// reportDBHealthOutcome counts the reasons that are only knowable after the
// collection has run. The deadline case is decided before, by the caller.
func reportDBHealthOutcome(h *domain.DatabaseHealth, onSkip dbHealthSkipFunc) {
	switch {
	case h == nil:
		onSkip(dbHealthSkipNoFields)
	case h.UnsupportedVersion:
		onSkip(dbHealthSkipUnsupported)
	case h.PrivilegeLimited:
		onSkip(dbHealthSkipPrivilege)
	}
}
