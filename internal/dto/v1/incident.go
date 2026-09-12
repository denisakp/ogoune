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
	// Explanation and HostEvents are populated on the detail path only, and both
	// are `omitempty` (spec 091).
	//
	// The omitempty is load-bearing, not tidiness: mapIncidentResponse is shared
	// with GET /api/v1/incidents, so without it every row of every incident
	// listing would gain two null keys -- a body change to an endpoint this
	// feature does not touch. With it, the listing and a detail response with
	// nothing to say both stay byte-identical to their pre-feature form
	// (FR-012's discipline, applied to the API).
	Explanation *IncidentExplanationResponse `json:"explanation,omitempty"`
	HostEvents  []HostEventResponse          `json:"host_events,omitempty"`
	// HostLink says which machine the surfaces above describe and where that
	// answer came from (spec 092). `omitempty` for the same reason as its
	// neighbours: this struct is shared with the list endpoint.
	HostLink   *HostLinkResponse `json:"host_link,omitempty"`
	StartedAt  string            `json:"started_at"`
	ResolvedAt *string           `json:"resolved_at"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
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

// IncidentExplanationResponse links what a check observed to what the kernel
// reported on the same host at nearly the same time (spec 091). Absent whenever
// no event fell in the incident's window.
//
// It states a co-occurrence and nothing more. There is deliberately no
// `caused_by`, no score, no confidence and no probability: the wording in `text`
// may read as an explanation, because a human knows a sentence is an
// interpretation, but no machine-readable field here asserts a mechanism. Both
// timestamps are always present so a reader can overrule the sentence.
//
// Nothing here is stored. It is recomputed from the incident and the events on
// every read, which is why improving the wording improves every past incident --
// and why the same incident read twice always reads the same.
// @name IncidentExplanationResponse
type IncidentExplanationResponse struct {
	// Text is the rendered sentence.
	//
	// The API is itself a renderer, which is why the prose is served rather than
	// left to the caller: the alternative was re-implementing the wording in
	// TypeScript for the SPA, and two implementations of one sentence drift.
	Text string `json:"text"`
	// HostID and HostName identify the machine whose kernel reported the event.
	HostID   string `json:"host_id"`
	HostName string `json:"host_name"`
	// IncidentAt is when the failure was confirmed; Cause is what the check saw.
	IncidentAt string `json:"incident_at"`
	Cause      string `json:"cause"`
	// Event is the one the sentence names. Its kind is always one this version
	// can phrase; unrecognised kinds appear in host_events and count toward
	// other_events, but are never named.
	Event HostEventResponse `json:"event"`
	// Precedes says whether the event happened at or before the failure. A
	// comparison of two timestamps, not a causal claim.
	Precedes bool `json:"precedes"`
	// OtherEvents counts the OTHER events in the window, unrecognised kinds
	// included. It counts events, not kernel reports: event.occurrences answers
	// that separate question.
	OtherEvents int `json:"other_events"`
	// WindowFrom and WindowTo are the incident host-context window, unchanged.
	// WindowTo may be earlier than the nominal end while the window is still
	// elapsing.
	WindowFrom string `json:"window_from"`
	WindowTo   string `json:"window_to"`
}

// HostLinkResponse says which machine an incident's host context and
// explanation describe, and how that is known (spec 092). Absent when the
// incident has no machine to describe.
//
// An incident is a historical record. The machine it happened on used to be
// resolved through the monitor at read time, so moving a monitor rewrote what
// its past incidents displayed. It is recorded when the incident opens now;
// incidents from before that fall back to the monitor's current machine and
// say so here, so a reader can tell a record from an inference.
// @name HostLinkResponse
type HostLinkResponse struct {
	HostID string `json:"host_id"`
	// Source is "recorded" -- written when the incident opened, authoritative --
	// or "inferred" -- the monitor's machine today, for an incident created
	// before Ogoune recorded it. An inferred machine may not be the one that
	// was involved. Treat an unknown value as "inferred".
	Source string `json:"source"`
	// Exists is false when that machine has since been deleted. The record
	// stands; its name and metrics can no longer be shown. A deleted recorded
	// machine is still "recorded" -- it does not fall back.
	Exists bool `json:"exists"`
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
