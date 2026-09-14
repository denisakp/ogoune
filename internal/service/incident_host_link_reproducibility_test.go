package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
)

// countByMachine groups incidents by the machine each one resolves to. This is
// the shape any historical report would take, and the question spec 092 US3
// asks is whether the answer holds still while monitors move.
func countByMachine(incidents []*domain.Incident) map[string]int {
	out := map[string]int{}
	for _, inc := range incidents {
		id, src := inc.ResolveHost()
		if src == domain.HostLinkSourceNone {
			continue
		}
		out[id]++
	}
	return out
}

func moveMonitor(incidents []*domain.Incident, to string) {
	for _, inc := range incidents {
		t := to
		inc.Resource.HostID = &t
	}
}

// SC-006: count incidents per machine, move a monitor, count again -- identical
// for incidents created after this change.
func TestIncidentsPerMachine_IsReproducibleForRecordedIncidents(t *testing.T) {
	hostA, hostB := "host-A", "host-B"
	recorded := []*domain.Incident{
		{Base: domain.Base{ID: "r1"}, StartedAt: time.Now(), HostID: &hostA, HostLinkRecorded: true, Resource: domain.Resource{HostID: &hostA}},
		{Base: domain.Base{ID: "r2"}, StartedAt: time.Now(), HostID: &hostA, HostLinkRecorded: true, Resource: domain.Resource{HostID: &hostA}},
		{Base: domain.Base{ID: "r3"}, StartedAt: time.Now(), HostID: &hostB, HostLinkRecorded: true, Resource: domain.Resource{HostID: &hostB}},
	}

	before := countByMachine(recorded)
	moveMonitor(recorded, "host-Z")
	after := countByMachine(recorded)

	assert.Equal(t, before, after, "a recorded incident does not follow its monitor")
	assert.Equal(t, map[string]int{"host-A": 2, "host-B": 1}, after)
}

// The converse, pinned on purpose: incidents from before recording DO change
// under the same move. This is the limit the feature cannot cross -- nothing
// anywhere records what a monitor was attached to in the past -- and a test that
// says so is what stops someone later believing the back catalogue is
// reproducible.
func TestIncidentsPerMachine_IsNotReproducibleForPreRecordingIncidents(t *testing.T) {
	hostA := "host-A"
	old := []*domain.Incident{
		{Base: domain.Base{ID: "o1"}, StartedAt: time.Now(), Resource: domain.Resource{HostID: &hostA}},
		{Base: domain.Base{ID: "o2"}, StartedAt: time.Now(), Resource: domain.Resource{HostID: &hostA}},
	}

	before := countByMachine(old)
	moveMonitor(old, "host-Z")
	after := countByMachine(old)

	assert.Equal(t, map[string]int{"host-A": 2}, before)
	assert.Equal(t, map[string]int{"host-Z": 2}, after,
		"pre-recording incidents follow the monitor, permanently; that is a property of the data, not a defect here")
	assert.NotEqual(t, before, after)
}

// The two populations side by side, as a real database will hold them for years.
func TestIncidentsPerMachine_MixedPopulationChangesOnlyInItsOldHalf(t *testing.T) {
	hostA := "host-A"
	mixed := []*domain.Incident{
		{Base: domain.Base{ID: "r"}, StartedAt: time.Now(), HostID: &hostA, HostLinkRecorded: true, Resource: domain.Resource{HostID: &hostA}},
		{Base: domain.Base{ID: "o"}, StartedAt: time.Now(), Resource: domain.Resource{HostID: &hostA}},
	}

	require.Equal(t, map[string]int{"host-A": 2}, countByMachine(mixed))
	moveMonitor(mixed, "host-Z")
	assert.Equal(t, map[string]int{"host-A": 1, "host-Z": 1}, countByMachine(mixed),
		"the recorded one stays, the old one moves")
}

var _ = context.Background // keep the import stable if helpers grow
