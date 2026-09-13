package monitoring

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

// failingHosts wraps the fake to make one lookup fail: the negative path the
// freeze rule must survive without blocking creation.
type failingHosts struct {
	port.HostRepository
	err error
}

func (f failingHosts) FindByID(ctx context.Context, id string) (*domain.Host, error) {
	return nil, f.err
}

func capsResource(id string, hostID *string) *domain.Resource {
	return &domain.Resource{
		Base: domain.Base{ID: id}, Name: "svc",
		Target: "https://example.com", Type: domain.ResourceHTTP,
		Status: domain.StatusDown, Interval: 60, Timeout: 10, IsActive: true,
		HostID: hostID,
	}
}

func onlyIncident(t *testing.T, incidents *fake.IncidentFake, resourceID string) *domain.Incident {
	t.Helper()
	found, err := incidents.FindByResource(context.Background(), resourceID, 10, 0)
	require.NoError(t, err)
	require.Len(t, found, 1, "exactly one incident, whatever was or was not recorded")
	return found[0]
}

// The four-state table from spec 093, plus the negative path: a lookup that
// fails records "not known" and the incident is created anyway.
func TestCreateIncident_FreezesTheHostDeclaration(t *testing.T) {
	now := time.Now()
	hostID := "host-1"
	declared := &domain.HostCapabilities{
		Kmsg:      domain.Capability{Available: false, Reason: domain.CapabilityReasonUnreadable},
		CgroupOOM: domain.Capability{Available: true},
		Segfault:  domain.Capability{Available: false, Reason: domain.CapabilityReasonUnreadable},
	}

	cases := map[string]struct {
		resourceHost *string
		host         *domain.Host // nil: not created in the fake
		hostsRepo    func(*fake.HostFake) port.HostRepository
		wantState    domain.HostCapabilitiesState
		wantCopy     *domain.HostCapabilities
	}{
		"no machine on the monitor": {
			resourceHost: nil, wantState: domain.HostCapabilitiesNoMachine,
		},
		"host declared: copy frozen": {
			resourceHost: &hostID,
			host:         &domain.Host{Base: domain.Base{ID: hostID}, Name: "h", LastSeenAt: &now, Capabilities: declared, CapabilitiesAt: &now},
			wantState:    domain.HostCapabilitiesDeclared, wantCopy: declared,
		},
		"host connected, no declaration (older agent)": {
			resourceHost: &hostID,
			host:         &domain.Host{Base: domain.Base{ID: hostID}, Name: "h", LastSeenAt: &now},
			wantState:    domain.HostCapabilitiesNotReported,
		},
		"host never connected": {
			resourceHost: &hostID,
			host:         &domain.Host{Base: domain.Base{ID: hostID}, Name: "h"},
			wantState:    domain.HostCapabilitiesNotKnown,
		},
		"host lookup fails: not known, incident still created": {
			resourceHost: &hostID,
			hostsRepo:    func(f *fake.HostFake) port.HostRepository { return failingHosts{f, errors.New("db down")} },
			wantState:    domain.HostCapabilitiesNotKnown,
		},
		"host missing from the repository: not known": {
			resourceHost: &hostID, wantState: domain.HostCapabilitiesNotKnown,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			svc, incidents, _, _, _, asynqClient := setupTestService()
			defer asynqClient.Close()
			hosts := fake.NewHostFake()
			if tc.host != nil {
				require.NoError(t, hosts.Create(context.Background(), tc.host))
			}
			var repo port.HostRepository = hosts
			if tc.hostsRepo != nil {
				repo = tc.hostsRepo(hosts)
			}
			svc.WithHosts(repo)

			resource := capsResource("res-caps-"+name, tc.resourceHost)
			require.NoError(t, svc.CreateIncident(context.Background(), resource,
				domain.CheckResult{Status: "down", ResponseData: "timeout"}))

			inc := onlyIncident(t, incidents, resource.ID)
			require.NotNil(t, inc.HostCapabilitiesState, "a new incident always records a state")
			assert.Equal(t, tc.wantState, *inc.HostCapabilitiesState)
			assert.Equal(t, tc.wantState, inc.FrozenCapabilitiesState())
			if tc.wantCopy == nil {
				assert.Nil(t, inc.HostCapabilities)
			} else {
				require.NotNil(t, inc.HostCapabilities)
				assert.Equal(t, *tc.wantCopy, *inc.HostCapabilities)
				assert.NotSame(t, tc.wantCopy, inc.HostCapabilities, "a copy, not the host's pointer")
			}
		})
	}
}

// The wiring an older bootstrap would produce: no host repository attached.
// Creation must neither panic nor pretend -- "not known" is what was recorded.
func TestCreateIncident_WithoutHostsWiring_RecordsNotKnown(t *testing.T) {
	svc, incidents, _, _, _, asynqClient := setupTestService()
	defer asynqClient.Close()
	hostID := "host-1"

	resource := capsResource("res-caps-unwired", &hostID)
	require.NoError(t, svc.CreateIncident(context.Background(), resource,
		domain.CheckResult{Status: "down", ResponseData: "timeout"}))

	inc := onlyIncident(t, incidents, resource.ID)
	assert.Equal(t, domain.HostCapabilitiesNotKnown, inc.FrozenCapabilitiesState())
	assert.Nil(t, inc.HostCapabilities)
}

// The copy is frozen: the host's declaration changing afterwards does not
// reach the incident. This is the 092 rule applied to the declaration.
func TestCreateIncident_FrozenCopyDoesNotFollowTheHost(t *testing.T) {
	svc, incidents, _, _, _, asynqClient := setupTestService()
	defer asynqClient.Close()
	now := time.Now()
	hostID := "host-1"
	hosts := fake.NewHostFake()
	host := &domain.Host{Base: domain.Base{ID: hostID}, Name: "h", LastSeenAt: &now,
		Capabilities:   &domain.HostCapabilities{Kmsg: domain.Capability{Available: false, Reason: domain.CapabilityReasonUnreadable}, CgroupOOM: domain.Capability{Available: true}, Segfault: domain.Capability{Available: false, Reason: domain.CapabilityReasonUnreadable}},
		CapabilitiesAt: &now}
	require.NoError(t, hosts.Create(context.Background(), host))
	svc.WithHosts(hosts)

	resource := capsResource("res-caps-frozen", &hostID)
	require.NoError(t, svc.CreateIncident(context.Background(), resource,
		domain.CheckResult{Status: "down", ResponseData: "timeout"}))

	// The container is recreated with kernel-log access.
	host.Capabilities = &domain.HostCapabilities{Kmsg: domain.Capability{Available: true}, CgroupOOM: domain.Capability{Available: true}, Segfault: domain.Capability{Available: true}}
	require.NoError(t, hosts.UpdateSnapshot(context.Background(), host))

	inc := onlyIncident(t, incidents, resource.ID)
	require.NotNil(t, inc.HostCapabilities)
	assert.False(t, inc.HostCapabilities.Kmsg.Available, "what it could observe THEN")
}
