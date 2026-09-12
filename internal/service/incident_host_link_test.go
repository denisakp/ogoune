package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

// The defect spec 092 removes, as a test: the host context describes the machine
// recorded when the incident opened, whatever the monitor points at today.
func TestGetIncidentByID_HostContextUsesTheRecordedMachine(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	ctx := context.Background()
	started := f.now.Add(-2 * time.Minute)

	// Two machines. The monitor is attached to B today; the incident was
	// recorded on A.
	require.NoError(t, f.hosts.Create(ctx, &domain.Host{Base: domain.Base{ID: "host-A"}, Name: "web-A"}))
	require.NoError(t, f.hosts.Create(ctx, &domain.Host{Base: domain.Base{ID: "host-B"}, Name: "web-B"}))
	hostA, hostB := "host-A", "host-B"

	inc := &domain.Incident{
		Base:       domain.Base{ID: "inc-moved"},
		ResourceID: "res-moved",
		Resource:   domain.Resource{Base: domain.Base{ID: "res-moved"}, Name: "svc", HostID: &hostB},
		StartedAt:  started,
		HostID:     &hostA, HostLinkRecorded: true,
	}
	_, err := f.incidents.Create(ctx, inc)
	require.NoError(t, err)

	// Only A has samples in the window. If the service asked about B it would
	// find nothing and return no context at all.
	f.seedSample(t, "host-A", started.Add(-30*time.Second), 91.0, 40.0, nil)

	got, err := f.svc.GetIncidentByID(ctx, inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.HostContext, "the recorded machine has samples; the current one does not")
	assert.Equal(t, "host-A", got.HostContext.HostID)
	assert.Equal(t, "web-A", got.HostContext.HostName)
}

// Move the monitor again: still A.
func TestGetIncidentByID_RecordedMachineSurvivesRepeatedMoves(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	ctx := context.Background()
	started := f.now.Add(-2 * time.Minute)
	require.NoError(t, f.hosts.Create(ctx, &domain.Host{Base: domain.Base{ID: "host-A"}, Name: "web-A"}))
	hostA := "host-A"

	inc := &domain.Incident{
		Base: domain.Base{ID: "inc-moves"}, ResourceID: "res-moves",
		Resource:  domain.Resource{Base: domain.Base{ID: "res-moves"}, Name: "svc"},
		StartedAt: started, HostID: &hostA, HostLinkRecorded: true,
	}
	_, err := f.incidents.Create(ctx, inc)
	require.NoError(t, err)
	f.seedSample(t, "host-A", started.Add(-30*time.Second), 50.0, 50.0, nil)

	for _, current := range []string{"host-B", "host-C", ""} {
		got, err := f.svc.GetIncidentByID(ctx, inc.ID)
		require.NoError(t, err)
		require.NotNil(t, got.HostContext, "monitor now on %q", current)
		assert.Equal(t, "host-A", got.HostContext.HostID, "monitor now on %q", current)
		// The fake returns a copy; mutate the stored resource for the next round.
		stored, _ := f.incidents.FindByID(ctx, inc.ID)
		if current == "" {
			stored.Resource.HostID = nil
		} else {
			c := current
			stored.Resource.HostID = &c
		}
	}
}

// A monitor that had no machine when the incident opened never falls back to
// one attached later (spec Q2, SC-005).
func TestGetIncidentByID_RecordedAbsenceNeverFallsBack(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	ctx := context.Background()
	started := f.now.Add(-2 * time.Minute)
	require.NoError(t, f.hosts.Create(ctx, &domain.Host{Base: domain.Base{ID: "host-later"}, Name: "web-later"}))
	later := "host-later"

	inc := &domain.Incident{
		Base: domain.Base{ID: "inc-absent"}, ResourceID: "res-absent",
		// Attached to a machine TODAY -- which has samples -- but recorded as
		// having had none when it opened.
		Resource:  domain.Resource{Base: domain.Base{ID: "res-absent"}, Name: "svc", HostID: &later},
		StartedAt: started, HostID: nil, HostLinkRecorded: true,
	}
	_, err := f.incidents.Create(ctx, inc)
	require.NoError(t, err)
	f.seedSample(t, "host-later", started.Add(-30*time.Second), 99.0, 99.0, nil)

	got, err := f.svc.GetIncidentByID(ctx, inc.ID)
	require.NoError(t, err)
	assert.Nil(t, got.HostContext, "the monitor genuinely had no machine; a machine attached later is not evidence")
}

// An incident from before recording falls back to the monitor's current machine
// -- and that is what keeps the back catalogue from going blank (US2).
func TestGetIncidentByID_PreRecordingIncidentFallsBack(t *testing.T) {
	f := newHostContextFixture(t, 7*24*time.Hour)
	ctx := context.Background()
	started := f.now.Add(-2 * time.Minute)
	require.NoError(t, f.hosts.Create(ctx, &domain.Host{Base: domain.Base{ID: "host-now"}, Name: "web-now"}))
	now := "host-now"

	inc := &domain.Incident{
		Base: domain.Base{ID: "inc-old"}, ResourceID: "res-old",
		Resource:  domain.Resource{Base: domain.Base{ID: "res-old"}, Name: "svc", HostID: &now},
		StartedAt: started,
		// Zero values: what a pre-migration row looks like.
	}
	_, err := f.incidents.Create(ctx, inc)
	require.NoError(t, err)
	f.seedSample(t, "host-now", started.Add(-30*time.Second), 60.0, 60.0, nil)

	got, err := f.svc.GetIncidentByID(ctx, inc.ID)
	require.NoError(t, err)
	require.NotNil(t, got.HostContext, "historical incidents keep their context rather than going blank")
	assert.Equal(t, "host-now", got.HostContext.HostID)
}
