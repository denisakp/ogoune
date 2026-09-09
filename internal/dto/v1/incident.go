package v1

import "github.com/denisakp/ogoune/internal/domain"

// IncidentResponse is the v1 API representation of an incident.
//
// It is a superset of the legacy root shape: it keeps the v1-native `monitor_id`
// + derived `status`, and also carries the rich fields the frontend renders
// (`resource_id`, the embedded `resource`, `details`, `event_steps`,
// `diagnostics`, `updated_at`) so the migration needs no frontend type change.
// The heavy fields (`resource`, `event_steps`, `diagnostics`) are populated on
// the detail endpoint; the list endpoint hydrates `resource` only.
// @name IncidentResponse
type IncidentResponse struct {
	ID          string                      `json:"id"`
	MonitorID   string                      `json:"monitor_id"`
	ResourceID  string                      `json:"resource_id"`
	Resource    domain.Resource             `json:"resource"`
	Cause       string                      `json:"cause"`
	Status      string                      `json:"status"` // "open" or "resolved"
	Details     string                      `json:"details"`
	EventSteps  []domain.IncidentEventStep  `json:"event_steps"`
	Diagnostics *domain.IncidentDiagnostics `json:"diagnostics"`
	HostContext *HostContextResponse        `json:"host_context"`
	StartedAt   string                      `json:"started_at"`
	ResolvedAt  *string                     `json:"resolved_at"`
	CreatedAt   string                      `json:"created_at"`
	UpdatedAt   string                      `json:"updated_at"`
}

// HostContextResponse is what the monitor's host was doing around the moment the
// incident opened (spec 089). Optional and nullable: absent whenever the monitor
// has no host attached, the host reported no samples in the window, the window
// predates retention, or the lookup failed. Absence is expressed by a null
// object, never by a zero-filled one.
//
// Owned by the v1 layer rather than embedded from the domain, so a domain change
// cannot silently alter the public contract.
type HostContextResponse struct {
	HostID     string  `json:"host_id"`
	HostName   string  `json:"host_name"`
	PeakCPUPct float64 `json:"peak_cpu_pct"`
	PeakMemPct float64 `json:"peak_mem_pct"`
	// WorstDisk is the highest-utilisation mount seen in the window, or null when
	// the host reported none. Its absence never suppresses the other figures.
	WorstDisk *WorstDiskResponse `json:"worst_disk"`
	// SampleCount is always >= 1 when this object is present, so a two-sample
	// aggregate is never mistaken for a full one.
	SampleCount int `json:"sample_count"`
	// Resolution is "full" or "reduced". Treat an unknown value as "reduced".
	Resolution string `json:"resolution"`
	WindowFrom string `json:"window_from"`
	// WindowTo may be earlier than the nominal window end while the window is
	// still elapsing on a fresh incident.
	WindowTo string `json:"window_to"`
}

// WorstDiskResponse is a single mount and its utilisation percentage.
type WorstDiskResponse struct {
	Mount   string  `json:"mount"`
	UsedPct float64 `json:"used_pct"`
}

// IncidentListFilters holds validated query parameters for listing incidents.
type IncidentListFilters struct {
	MonitorID string // optional ULID filter
	Status    string // "open", "resolved", or "" (all)
}
