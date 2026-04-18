package v1

// IncidentResponse is the v1 API representation of an incident.
// @name IncidentResponse
type IncidentResponse struct {
	ID         string  `json:"id"`
	MonitorID  string  `json:"monitor_id"`
	Cause      string  `json:"cause"`
	Status     string  `json:"status"` // "open" or "resolved"
	StartedAt  string  `json:"started_at"`
	ResolvedAt *string `json:"resolved_at"`
	CreatedAt  string  `json:"created_at"`
}

// IncidentListFilters holds validated query parameters for listing incidents.
type IncidentListFilters struct {
	MonitorID string // optional ULID filter
	Status    string // "open", "resolved", or "" (all)
}
