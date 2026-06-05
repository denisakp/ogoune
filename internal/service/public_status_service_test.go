package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

func mkResource(t *testing.T, id, name, componentID string, status domain.ResourceStatus) *domain.Resource {
	t.Helper()
	r := &domain.Resource{
		Name:     name,
		Target:   name + ".acme.com",
		Status:   status,
		IsActive: true,
	}
	r.ID = id
	if componentID != "" {
		cid := componentID
		r.ComponentID = &cid
	}
	return r
}

type stubMaintenances struct {
	activeForResource map[string]bool
}

func (s *stubMaintenances) Create(context.Context, *domain.Maintenance) (*domain.Maintenance, error) {
	return nil, nil
}
func (s *stubMaintenances) FindByID(context.Context, string) (*domain.Maintenance, error) {
	return nil, nil
}
func (s *stubMaintenances) List(context.Context, string, int, int) ([]*domain.Maintenance, error) {
	return nil, nil
}
func (s *stubMaintenances) Update(context.Context, *domain.Maintenance) error { return nil }
func (s *stubMaintenances) Delete(context.Context, string) error              { return nil }
func (s *stubMaintenances) FindActiveForResource(_ context.Context, resourceID string, _ time.Time) ([]*domain.Maintenance, error) {
	if s == nil || !s.activeForResource[resourceID] {
		return nil, nil
	}
	return []*domain.Maintenance{{}}, nil
}

func setup(t *testing.T) (*PublicStatusService, *fake.ResourceFake, *fake.ComponentFake, *fake.IncidentFake, *fake.UptimeDailyAggRepository, *stubMaintenances) {
	t.Helper()
	resources := fake.NewResourceFake()
	components := fake.NewComponentFake()
	incidents := fake.NewIncidentFake()
	maintenances := &stubMaintenances{activeForResource: map[string]bool{}}
	aggs := fake.NewUptimeDailyAggRepository()
	svc := NewPublicStatusService(resources, components, incidents, maintenances, aggs)
	svc.SetClock(func() time.Time { return time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC) })
	return svc, resources, components, incidents, aggs, maintenances
}

func TestPublicStatus_AllUp_VerdictOperational(t *testing.T) {
	svc, resources, components, _, _, _ := setup(t)
	c, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	_, _ = resources.Create(context.Background(), mkResource(t, "r1", "api1", c.ID, domain.StatusUp))
	_, _ = resources.Create(context.Background(), mkResource(t, "r2", "api2", c.ID, domain.StatusUp))

	out, err := svc.GetCurrent(context.Background())
	require.NoError(t, err)
	assert.Equal(t, dto.VerdictOperational, out.Verdict.Status)
	require.Len(t, out.Components, 1)
	assert.Equal(t, dto.PublicStateUp, out.Components[0].AggregatedState)
}

func TestPublicStatus_OneResourceDegraded_VerdictPartial(t *testing.T) {
	svc, resources, components, _, _, _ := setup(t)
	c, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	_, _ = resources.Create(context.Background(), mkResource(t, "r1", "api1", c.ID, domain.StatusUp))
	_, _ = resources.Create(context.Background(), mkResource(t, "r2", "api2", c.ID, domain.StatusFlapping))

	out, err := svc.GetCurrent(context.Background())
	require.NoError(t, err)
	assert.Equal(t, dto.VerdictPartialDegradation, out.Verdict.Status)
	assert.Equal(t, dto.PublicStateDegraded, out.Components[0].AggregatedState)
}

func TestPublicStatus_OneComponentFullyDown_VerdictMajor(t *testing.T) {
	svc, resources, components, _, _, _ := setup(t)
	c1, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	c2, _ := components.Create(context.Background(), &domain.Component{Name: "Web"})
	_, _ = resources.Create(context.Background(), mkResource(t, "r1", "api1", c1.ID, domain.StatusDown))
	_, _ = resources.Create(context.Background(), mkResource(t, "r2", "web1", c2.ID, domain.StatusUp))

	out, err := svc.GetCurrent(context.Background())
	require.NoError(t, err)
	assert.Equal(t, dto.VerdictMajorOutage, out.Verdict.Status)
}

func TestPublicStatus_HalfOfComponentsDegradedOrDown_VerdictMajor(t *testing.T) {
	svc, resources, components, _, _, _ := setup(t)
	// 4 components: 2 with a flapping resource → 50% degraded → major.
	for i, name := range []string{"A", "B", "C", "D"} {
		c, _ := components.Create(context.Background(), &domain.Component{Name: name})
		status := domain.StatusUp
		if i < 2 {
			status = domain.StatusFlapping
		}
		_, _ = resources.Create(context.Background(), mkResource(t, "r"+name, "res"+name, c.ID, status))
	}

	out, err := svc.GetCurrent(context.Background())
	require.NoError(t, err)
	assert.Equal(t, dto.VerdictMajorOutage, out.Verdict.Status)
}

func TestPublicStatus_MaintenanceOverridesDown(t *testing.T) {
	svc, resources, components, _, _, maint := setup(t)
	c, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	_, _ = resources.Create(context.Background(), mkResource(t, "r1", "api1", c.ID, domain.StatusDown))
	maint.activeForResource["r1"] = true

	out, err := svc.GetCurrent(context.Background())
	require.NoError(t, err)
	// Maintenance is not a down state for verdict purposes; the component
	// reports maintenance and the verdict is partial (non-up but not down).
	require.Len(t, out.Components, 1)
	assert.Equal(t, dto.PublicStateMaintenance, out.Components[0].Resources[0].CurrentState)
	assert.Equal(t, dto.VerdictPartialDegradation, out.Verdict.Status)
}

func TestPublicStatus_RibbonReadsFromAggregates(t *testing.T) {
	svc, resources, components, _, aggs, _ := setup(t)
	c, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	_, _ = resources.Create(context.Background(), mkResource(t, "r1", "api1", c.ID, domain.StatusUp))

	day := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	require.NoError(t, aggs.Upsert(context.Background(), &domain.UptimeDailyAgg{
		ResourceID: "r1", Day: day, UptimeRatio: 0.9876, Samples: 288, ComputedAt: day,
	}))

	out, err := svc.GetCurrent(context.Background())
	require.NoError(t, err)
	require.Len(t, out.Components, 1)
	ribbon := out.Components[0].Resources[0].UptimeRibbon
	require.Len(t, ribbon, 90)
	assert.Equal(t, "2026-06-04", ribbon[89].Day)
	assert.InDelta(t, 0.9876, ribbon[89].Ratio, 0.0001)
	// Missing days remain at 0 — aggregator backfills within 5 min in production.
	assert.Equal(t, 0.0, ribbon[0].Ratio)
}

// ---------- US2: GetIncidents ----------

func mkIncidentAt(t *testing.T, id, resourceID, cause string, start time.Time, resolvedAt *time.Time) *domain.Incident {
	t.Helper()
	inc := &domain.Incident{
		ResourceID: resourceID,
		Cause:      cause,
		StartedAt:  start,
		ResolvedAt: resolvedAt,
	}
	inc.ID = id
	return inc
}

func TestGetIncidents_GroupsByMonthNewestFirst(t *testing.T) {
	svc, resources, components, incidents, _, _ := setup(t)
	c, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	r := mkResource(t, "res-1", "api1", c.ID, domain.StatusUp)
	_, _ = resources.Create(context.Background(), r)

	resolved := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	_, _ = incidents.Create(context.Background(), mkIncidentAt(t, "i1", "res-1", "down A",
		time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC), &resolved))
	_, _ = incidents.Create(context.Background(), mkIncidentAt(t, "i2", "res-1", "down B",
		time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC), &resolved))
	_, _ = incidents.Create(context.Background(), mkIncidentAt(t, "i3", "res-1", "down C",
		time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC), &resolved))

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	out, err := svc.GetIncidents(context.Background(), from, to, "")
	require.NoError(t, err)
	assert.Equal(t, 3, out.Total)
	require.Len(t, out.Months, 3)
	// Newest month first.
	assert.Equal(t, "2026-06", out.Months[0].YearMonth)
	assert.Equal(t, "2026-05", out.Months[1].YearMonth)
	assert.Equal(t, "2026-04", out.Months[2].YearMonth)
}

func TestGetIncidents_FiltersByComponent(t *testing.T) {
	svc, resources, components, incidents, _, _ := setup(t)
	c1, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	c2, _ := components.Create(context.Background(), &domain.Component{Name: "Web"})
	_, _ = resources.Create(context.Background(), mkResource(t, "ra", "api1", c1.ID, domain.StatusUp))
	_, _ = resources.Create(context.Background(), mkResource(t, "rb", "web1", c2.ID, domain.StatusUp))

	resolved := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	_, _ = incidents.Create(context.Background(), mkIncidentAt(t, "i1", "ra", "api down",
		time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC), &resolved))
	_, _ = incidents.Create(context.Background(), mkIncidentAt(t, "i2", "rb", "web down",
		time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC), &resolved))

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	out, err := svc.GetIncidents(context.Background(), from, to, c1.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, out.Total)
	assert.Equal(t, "i1", out.Months[0].Incidents[0].ID)
}

func TestGetIncidents_EmptyResultStillReturnsEnvelope(t *testing.T) {
	svc, _, _, _, _, _ := setup(t)
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	out, err := svc.GetIncidents(context.Background(), from, to, "")
	require.NoError(t, err)
	assert.Equal(t, 0, out.Total)
	assert.Empty(t, out.Months)
}

func TestGetIncidents_FromAfterTo_Errors(t *testing.T) {
	svc, _, _, _, _, _ := setup(t)
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.GetIncidents(context.Background(), from, to, "")
	require.Error(t, err)
}

// ---------- US3: GetUptime ----------

func TestGetUptime_MultiMonthRange(t *testing.T) {
	svc, resources, components, _, aggs, _ := setup(t)
	c, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	_, _ = resources.Create(context.Background(), mkResource(t, "r1", "api1", c.ID, domain.StatusUp))

	for i := 0; i < 3; i++ {
		day := time.Date(2026, 4, 1+i, 0, 0, 0, 0, time.UTC)
		require.NoError(t, aggs.Upsert(context.Background(), &domain.UptimeDailyAgg{
			ResourceID: "r1", Day: day, UptimeRatio: 0.99, Samples: 100, ComputedAt: day,
		}))
	}

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	out, err := svc.GetUptime(context.Background(), "", from, to)
	require.NoError(t, err)
	// 3 months × ~30 days = 92 entries in window.
	require.GreaterOrEqual(t, len(out.Days), 90)
	// First 3 known days should match.
	knownCount := 0
	for _, d := range out.Days {
		if d.Samples > 0 {
			knownCount++
			assert.InDelta(t, 0.99, d.UptimeRatio, 0.0001)
		}
	}
	assert.Equal(t, 3, knownCount)
}

func TestGetUptime_ComponentFilterScopesRatios(t *testing.T) {
	svc, resources, components, _, aggs, _ := setup(t)
	c1, _ := components.Create(context.Background(), &domain.Component{Name: "API"})
	c2, _ := components.Create(context.Background(), &domain.Component{Name: "Web"})
	_, _ = resources.Create(context.Background(), mkResource(t, "r-api", "api1", c1.ID, domain.StatusUp))
	_, _ = resources.Create(context.Background(), mkResource(t, "r-web", "web1", c2.ID, domain.StatusUp))

	day := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, aggs.Upsert(context.Background(), &domain.UptimeDailyAgg{
		ResourceID: "r-api", Day: day, UptimeRatio: 1.0, Samples: 100, ComputedAt: day,
	}))
	require.NoError(t, aggs.Upsert(context.Background(), &domain.UptimeDailyAgg{
		ResourceID: "r-web", Day: day, UptimeRatio: 0.5, Samples: 100, ComputedAt: day,
	}))

	// Scoped to c1 (API) — only r-api counts → ratio 1.0.
	out, err := svc.GetUptime(context.Background(), c1.ID, day, day)
	require.NoError(t, err)
	require.Len(t, out.Days, 1)
	assert.InDelta(t, 1.0, out.Days[0].UptimeRatio, 0.0001)
	assert.Equal(t, 100, out.Days[0].Samples)

	// No filter — mean of both → 0.75.
	out2, err := svc.GetUptime(context.Background(), "", day, day)
	require.NoError(t, err)
	assert.InDelta(t, 0.75, out2.Days[0].UptimeRatio, 0.0001)
}

func TestGetUptime_FromAfterTo_Errors(t *testing.T) {
	svc, _, _, _, _, _ := setup(t)
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.GetUptime(context.Background(), "", from, to)
	require.Error(t, err)
}
