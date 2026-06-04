package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

type EscalationRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewEscalationRepositorySQLC(rt SqlcRuntime) port.EscalationRepository {
	r := &EscalationRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *EscalationRepositorySQLC) unconfigured() error {
	return fmt.Errorf("escalation_repository_sqlc: unconfigured runtime")
}

func (r *EscalationRepositorySQLC) Create(ctx context.Context, p *domain.EscalationPolicy) error {
	p.EnsureID()
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
	switch {
	case r.pgQ != nil:
		if err := r.pgQ.CreateEscalationPolicy(ctx, pgsqlc.CreateEscalationPolicyParams{
			ID:         p.ID,
			Name:       p.Name,
			ScopeKind:  string(p.Scope.Kind),
			ScopeValue: p.Scope.Value,
			IsActive:   p.IsActive,
			Priority:   int32(p.Priority),
			CreatedAt:  pgtype.Timestamptz{Time: p.CreatedAt, Valid: true},
			UpdatedAt:  pgtype.Timestamptz{Time: p.UpdatedAt, Valid: true},
		}); err != nil {
			return fmt.Errorf("sqlc: create escalation policy: %w", err)
		}
		return r.insertStepsPG(ctx, p)
	case r.sqliteQ != nil:
		if err := r.sqliteQ.CreateEscalationPolicy(ctx, sqlitesqlc.CreateEscalationPolicyParams{
			ID:         p.ID,
			Name:       p.Name,
			ScopeKind:  string(p.Scope.Kind),
			ScopeValue: p.Scope.Value,
			IsActive:   boolToInt64(p.IsActive),
			Priority:   int64(p.Priority),
			CreatedAt:  p.CreatedAt,
			UpdatedAt:  p.UpdatedAt,
		}); err != nil {
			return fmt.Errorf("sqlc: create escalation policy: %w", err)
		}
		return r.insertStepsSQLite(ctx, p)
	default:
		return r.unconfigured()
	}
}

func (r *EscalationRepositorySQLC) insertStepsPG(ctx context.Context, p *domain.EscalationPolicy) error {
	for i, st := range p.Steps {
		if st.ID == "" {
			(&domain.Base{}).EnsureID()
			tmp := &domain.Base{}
			tmp.EnsureID()
			st.ID = tmp.ID
			p.Steps[i].ID = st.ID
		}
		raw, _ := json.Marshal(st.ChannelIDs)
		if err := r.pgQ.CreateEscalationStep(ctx, pgsqlc.CreateEscalationStepParams{
			ID:           st.ID,
			PolicyID:     p.ID,
			StepOrder:    int32(st.StepOrder),
			DelayMinutes: int32(st.DelayMinutes),
			ChannelIds:   raw,
		}); err != nil {
			return fmt.Errorf("sqlc: create escalation step: %w", err)
		}
	}
	return nil
}

func (r *EscalationRepositorySQLC) insertStepsSQLite(ctx context.Context, p *domain.EscalationPolicy) error {
	for i, st := range p.Steps {
		if st.ID == "" {
			tmp := &domain.Base{}
			tmp.EnsureID()
			st.ID = tmp.ID
			p.Steps[i].ID = st.ID
		}
		raw, _ := json.Marshal(st.ChannelIDs)
		if err := r.sqliteQ.CreateEscalationStep(ctx, sqlitesqlc.CreateEscalationStepParams{
			ID:           st.ID,
			PolicyID:     p.ID,
			StepOrder:    int64(st.StepOrder),
			DelayMinutes: int64(st.DelayMinutes),
			ChannelIds:   string(raw),
		}); err != nil {
			return fmt.Errorf("sqlc: create escalation step: %w", err)
		}
	}
	return nil
}

func (r *EscalationRepositorySQLC) FindByID(ctx context.Context, id string) (*domain.EscalationPolicy, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.FindEscalationPolicyByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find escalation policy: %w", err)
		}
		out := escalationFromPG(row)
		steps, err := r.pgQ.ListEscalationStepsByPolicy(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list steps: %w", err)
		}
		out.Steps = stepsFromPG(steps)
		return out, nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.FindEscalationPolicyByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, fmt.Errorf("sqlc: find escalation policy: %w", err)
		}
		out := escalationFromSQLite(row)
		steps, err := r.sqliteQ.ListEscalationStepsByPolicy(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list steps: %w", err)
		}
		out.Steps = stepsFromSQLite(steps)
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *EscalationRepositorySQLC) List(ctx context.Context) ([]*domain.EscalationPolicy, error) {
	switch {
	case r.pgQ != nil:
		rows, err := r.pgQ.ListEscalationPolicies(ctx)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list policies: %w", err)
		}
		out := make([]*domain.EscalationPolicy, len(rows))
		for i, row := range rows {
			p := escalationFromPG(row)
			steps, err := r.pgQ.ListEscalationStepsByPolicy(ctx, p.ID)
			if err != nil {
				return nil, fmt.Errorf("sqlc: list steps: %w", err)
			}
			p.Steps = stepsFromPG(steps)
			out[i] = p
		}
		return out, nil
	case r.sqliteQ != nil:
		rows, err := r.sqliteQ.ListEscalationPolicies(ctx)
		if err != nil {
			return nil, fmt.Errorf("sqlc: list policies: %w", err)
		}
		out := make([]*domain.EscalationPolicy, len(rows))
		for i, row := range rows {
			p := escalationFromSQLite(row)
			steps, err := r.sqliteQ.ListEscalationStepsByPolicy(ctx, p.ID)
			if err != nil {
				return nil, fmt.Errorf("sqlc: list steps: %w", err)
			}
			p.Steps = stepsFromSQLite(steps)
			out[i] = p
		}
		return out, nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *EscalationRepositorySQLC) Update(ctx context.Context, p *domain.EscalationPolicy) error {
	p.UpdatedAt = time.Now()
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.UpdateEscalationPolicy(ctx, pgsqlc.UpdateEscalationPolicyParams{
			ID:         p.ID,
			Name:       p.Name,
			ScopeKind:  string(p.Scope.Kind),
			ScopeValue: p.Scope.Value,
			IsActive:   p.IsActive,
			Priority:   int32(p.Priority),
			UpdatedAt:  pgtype.Timestamptz{Time: p.UpdatedAt, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("sqlc: update policy: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		if err := r.pgQ.DeleteEscalationStepsByPolicy(ctx, p.ID); err != nil {
			return fmt.Errorf("sqlc: delete old steps: %w", err)
		}
		return r.insertStepsPG(ctx, p)
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.UpdateEscalationPolicy(ctx, sqlitesqlc.UpdateEscalationPolicyParams{
			ID:         p.ID,
			Name:       p.Name,
			ScopeKind:  string(p.Scope.Kind),
			ScopeValue: p.Scope.Value,
			IsActive:   boolToInt64(p.IsActive),
			Priority:   int64(p.Priority),
			UpdatedAt:  p.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("sqlc: update policy: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		if err := r.sqliteQ.DeleteEscalationStepsByPolicy(ctx, p.ID); err != nil {
			return fmt.Errorf("sqlc: delete old steps: %w", err)
		}
		return r.insertStepsSQLite(ctx, p)
	default:
		return r.unconfigured()
	}
}

func (r *EscalationRepositorySQLC) Delete(ctx context.Context, id string) error {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.DeleteEscalationPolicy(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete policy: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.DeleteEscalationPolicy(ctx, id)
		if err != nil {
			return fmt.Errorf("sqlc: delete policy: %w", err)
		}
		if n == 0 {
			return repository.ErrNotFound
		}
		return nil
	default:
		return r.unconfigured()
	}
}

// Reorder reassigns priorities 1..N in the order given.
// Two-phase: temporarily shift each entry to a high "scratch" range to avoid
// transiently colliding with another active row, then assign final values.
// This works under either the PG partial-unique index or the SQLite triggers.
func (r *EscalationRepositorySQLC) Reorder(ctx context.Context, order []string) error {
	now := time.Now()
	const scratchBase = 100000
	switch {
	case r.pgQ != nil:
		for i, id := range order {
			if _, err := r.pgQ.SetEscalationPolicyPriority(ctx, pgsqlc.SetEscalationPolicyPriorityParams{
				ID:        id,
				Priority:  int32(scratchBase + i),
				UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			}); err != nil {
				return fmt.Errorf("sqlc: reorder phase1: %w", err)
			}
		}
		for i, id := range order {
			n, err := r.pgQ.SetEscalationPolicyPriority(ctx, pgsqlc.SetEscalationPolicyPriorityParams{
				ID:        id,
				Priority:  int32(i + 1),
				UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			})
			if err != nil {
				return fmt.Errorf("sqlc: reorder phase2: %w", err)
			}
			if n == 0 {
				return repository.ErrNotFound
			}
		}
		return nil
	case r.sqliteQ != nil:
		for i, id := range order {
			if _, err := r.sqliteQ.SetEscalationPolicyPriority(ctx, sqlitesqlc.SetEscalationPolicyPriorityParams{
				ID:        id,
				Priority:  int64(scratchBase + i),
				UpdatedAt: now,
			}); err != nil {
				return fmt.Errorf("sqlc: reorder phase1: %w", err)
			}
		}
		for i, id := range order {
			n, err := r.sqliteQ.SetEscalationPolicyPriority(ctx, sqlitesqlc.SetEscalationPolicyPriorityParams{
				ID:        id,
				Priority:  int64(i + 1),
				UpdatedAt: now,
			})
			if err != nil {
				return fmt.Errorf("sqlc: reorder phase2: %w", err)
			}
			if n == 0 {
				return repository.ErrNotFound
			}
		}
		return nil
	default:
		return r.unconfigured()
	}
}

func (r *EscalationRepositorySQLC) NextPriority(ctx context.Context) (int, error) {
	switch {
	case r.pgQ != nil:
		n, err := r.pgQ.NextEscalationPriority(ctx)
		if err != nil {
			return 0, fmt.Errorf("sqlc: next priority: %w", err)
		}
		return int(n), nil
	case r.sqliteQ != nil:
		n, err := r.sqliteQ.NextEscalationPriority(ctx)
		if err != nil {
			return 0, fmt.Errorf("sqlc: next priority: %w", err)
		}
		return int(n), nil
	default:
		return 0, r.unconfigured()
	}
}

// ---------- mapping ----------

func escalationFromPG(row pgsqlc.EscalationPolicy) *domain.EscalationPolicy {
	return &domain.EscalationPolicy{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Name: row.Name,
		Scope: domain.EscalationScope{
			Kind:  domain.EscalationScopeKind(row.ScopeKind),
			Value: row.ScopeValue,
		},
		IsActive: row.IsActive,
		Priority: int(row.Priority),
	}
}

func escalationFromSQLite(row sqlitesqlc.EscalationPolicy) *domain.EscalationPolicy {
	return &domain.EscalationPolicy{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Name: row.Name,
		Scope: domain.EscalationScope{
			Kind:  domain.EscalationScopeKind(row.ScopeKind),
			Value: row.ScopeValue,
		},
		IsActive: row.IsActive != 0,
		Priority: int(row.Priority),
	}
}

func stepsFromPG(rows []pgsqlc.EscalationStep) []domain.EscalationStep {
	out := make([]domain.EscalationStep, len(rows))
	for i, row := range rows {
		var ids []string
		_ = json.Unmarshal(row.ChannelIds, &ids)
		out[i] = domain.EscalationStep{
			ID:           row.ID,
			PolicyID:     row.PolicyID,
			StepOrder:    int(row.StepOrder),
			DelayMinutes: int(row.DelayMinutes),
			ChannelIDs:   ids,
		}
	}
	return out
}

func stepsFromSQLite(rows []sqlitesqlc.EscalationStep) []domain.EscalationStep {
	out := make([]domain.EscalationStep, len(rows))
	for i, row := range rows {
		var ids []string
		_ = json.Unmarshal([]byte(row.ChannelIds), &ids)
		out[i] = domain.EscalationStep{
			ID:           row.ID,
			PolicyID:     row.PolicyID,
			StepOrder:    int(row.StepOrder),
			DelayMinutes: int(row.DelayMinutes),
			ChannelIDs:   ids,
		}
	}
	return out
}
