package dto

import "time"

// StatusPageData represents the complete data structure for rendering the public status page.
type StatusPageData struct {
	Resources []ResourceStatusInfo `json:"resources"`
	Incidents []IncidentSummary    `json:"incidents"`
	Generated time.Time            `json:"generated"`
}

// ResourceStatusInfo represents a monitored resource's current status and uptime metrics.
type ResourceStatusInfo struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	CurrentStatus    string    `json:"current_status"`
	UptimeLast30Days float64   `json:"uptime_last_30_days"`
	LastChecked      time.Time `json:"last_checked"`
	ResponseTime     int       `json:"response_time_ms,omitempty"`
}

// IncidentSummary represents a high-level summary of an incident for public display.
type IncidentSummary struct {
	ID         string     `json:"id"`
	ResourceID string     `json:"resource_id"`
	Resource   string     `json:"resource"`
	Reason     string     `json:"reason"`
	Cause      string     `json:"cause"`
	StartedAt  time.Time  `json:"started_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	Duration   string     `json:"duration"`
	IsOngoing  bool       `json:"is_ongoing"`
}
