package monitoring

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

// The machine a monitor is attached to is frozen onto the incident when it
// opens (spec 092). Until now it was resolved through the monitor at read time,
// so moving a monitor rewrote what its past incidents displayed.
func TestCreateIncident_RecordsTheMachineAtCreation(t *testing.T) {
	svc, incidents, _, _, _, asynqClient := setupTestService()
	defer asynqClient.Close()

	hostA := "host-A"
	resource := &domain.Resource{
		Base: domain.Base{ID: "res-hl-recorded"}, Name: "svc",
		Target: "https://example.com", Type: domain.ResourceHTTP,
		Status: domain.StatusDown, Interval: 60, Timeout: 10, IsActive: true,
		HostID: &hostA,
	}

	require.NoError(t, svc.CreateIncident(context.Background(), resource,
		domain.CheckResult{Status: "down", ResponseData: "timeout"}))

	found, err := incidents.FindByResource(context.Background(), resource.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, found, 1)

	inc := found[0]
	require.NotNil(t, inc.HostID)
	assert.Equal(t, "host-A", *inc.HostID)
	assert.True(t, inc.HostLinkRecorded)

	// And the record does not follow the monitor.
	hostB := "host-B"
	resource.HostID = &hostB
	got, src := inc.ResolveHost()
	assert.Equal(t, "host-A", got)
	assert.Equal(t, domain.HostLinkSourceRecorded, src)
}

// "We looked, and there was none" is a fact, and it is the one that must never
// fall back to a machine attached later (spec Q2).
func TestCreateIncident_RecordsTheAbsenceOfAMachine(t *testing.T) {
	svc, incidents, _, _, _, asynqClient := setupTestService()
	defer asynqClient.Close()

	resource := &domain.Resource{
		Base: domain.Base{ID: "res-hl-none"}, Name: "svc",
		Target: "https://example.com", Type: domain.ResourceHTTP,
		Status: domain.StatusDown, Interval: 60, Timeout: 10, IsActive: true,
		HostID: nil,
	}

	require.NoError(t, svc.CreateIncident(context.Background(), resource,
		domain.CheckResult{Status: "down", ResponseData: "timeout"}))

	found, err := incidents.FindByResource(context.Background(), resource.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, found, 1)

	inc := found[0]
	assert.Nil(t, inc.HostID)
	assert.True(t, inc.HostLinkRecorded, "absence is recorded, not left as the pre-feature zero value")

	// Attach a machine afterwards: the incident must not pick it up.
	later := "host-later"
	inc.Resource.HostID = &later
	_, src := inc.ResolveHost()
	assert.Equal(t, domain.HostLinkSourceNone, src,
		"a monitor that had no machine when the incident opened must never fall back to one attached later")
}

// FR-004: recording must not be able to prevent an incident from being created.
// There is no separate step that could fail -- the machine is two fields on the
// same insert -- so the property is structural. This pins it: creation
// behaves identically with and without a machine to record.
func TestCreateIncident_RecordingCannotBlockCreation(t *testing.T) {
	for name, hostID := range map[string]*string{
		"with a machine":    func() *string { s := "host-A"; return &s }(),
		"without a machine": nil,
	} {
		t.Run(name, func(t *testing.T) {
			svc, incidents, _, _, _, asynqClient := setupTestService()
			defer asynqClient.Close()

			resource := &domain.Resource{
				Base: domain.Base{ID: "res-hl-" + name}, Name: "svc",
				Target: "https://example.com", Type: domain.ResourceHTTP,
				Status: domain.StatusDown, Interval: 60, Timeout: 10, IsActive: true,
				HostID: hostID,
			}
			require.NoError(t, svc.CreateIncident(context.Background(), resource,
				domain.CheckResult{Status: "down", ResponseData: "timeout"}))

			found, err := incidents.FindByResource(context.Background(), resource.ID, 10, 0)
			require.NoError(t, err)
			assert.Len(t, found, 1, "exactly one incident, whatever there was to record")
		})
	}
}

// FR-010 / FR-012: the recorded machine is a fact written next to the incident.
// It takes no part in creation, duplicate suppression, resolution or the flap
// notifications -- the event-step sequence is the same with and without it.
func TestIncidentLifecycle_IdenticalWithAndWithoutRecordedMachine(t *testing.T) {
	hostA := "host-A"
	run := func(t *testing.T, id string, hostID *string) []domain.IncidentEventStepType {
		svc, incidents, steps, _, _, asynqClient := setupTestService()
		defer asynqClient.Close()
		ctx := context.Background()

		resource := &domain.Resource{
			Base: domain.Base{ID: id}, Name: "svc",
			Target: "https://example.com", Type: domain.ResourceHTTP,
			Status: domain.StatusDown, Interval: 60, Timeout: 10, IsActive: true,
			HostID: hostID,
		}
		down := domain.CheckResult{Status: "down", ResponseData: "connection timeout"}
		up := domain.CheckResult{Status: "up", ResponseData: "OK"}

		require.NoError(t, svc.CreateIncident(ctx, resource, down))
		require.NoError(t, svc.CreateIncident(ctx, resource, down), "second confirmation is suppressed, not an error")
		require.NoError(t, svc.NotifyFlapping(ctx, resource, 4, 300, 30))
		require.NoError(t, svc.ResolveIncident(ctx, resource, up))
		require.NoError(t, svc.NotifyStabilized(ctx, resource, domain.StatusUp))

		found, err := incidents.FindByResource(ctx, resource.ID, 10, 0)
		require.NoError(t, err)
		require.Len(t, found, 1, "duplicate prevention holds regardless of the link")
		require.NotNil(t, found[0].ResolvedAt)

		all, err := steps.List(ctx, 100, 0)
		require.NoError(t, err)
		// The fake iterates a map; order by time, then by name for ties.
		sort.Slice(all, func(i, j int) bool {
			if !all[i].CreatedAt.Equal(all[j].CreatedAt) {
				return all[i].CreatedAt.Before(all[j].CreatedAt)
			}
			return all[i].Step < all[j].Step
		})
		seq := make([]domain.IncidentEventStepType, 0, len(all))
		for _, s := range all {
			seq = append(seq, s.Step)
		}
		return seq
	}

	with := run(t, "res-lifecycle-with", &hostA)
	without := run(t, "res-lifecycle-without", nil)

	require.NotEmpty(t, with)
	assert.Equal(t, with, without, "the link changes what the incident remembers, never what it does")
}
