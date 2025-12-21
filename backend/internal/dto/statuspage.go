package dto

import "time"

// StatusPageData represents the complete data structure for rendering the public status page.
type StatusPageData struct {
	GlobalStatus string               `json:"global_status"`
	GeneratedAt  time.Time            `json:"generated_at"`
	Resources    []ResourceStatusInfo `json:"resources"`
}

// ResourceStatusInfo represents a monitored resource's current status and uptime metrics.
type ResourceStatusInfo struct {
	ID                         string   `json:"id"`
	Name                       string   `json:"name"`
	CurrentStatus              string   `json:"current_status"`
	UptimePercentageLast90Days float64  `json:"uptime_percentage_last_90_days"`
	DailyStatusLast90Days      []string `json:"daily_status_last_90_days"`
}

// IncidentSummary represents a high-level summary of an incident for public display.
type IncidentSummary struct {
	ID         string     `json:"id"`
	ResourceID string     `json:"resource_id"`
	Resource   string     `json:"resource"`
	Cause      string     `json:"cause"`
	StartedAt  time.Time  `json:"started_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	Duration   string     `json:"duration"`
	IsOngoing  bool       `json:"is_ongoing"`
}

// ResourceDetailStatusData represents detailed status information for a single resource
type ResourceDetailStatusData struct {
	ID                    string              `json:"id"`
	Name                  string              `json:"name"`
	CurrentStatus         string              `json:"current_status"`
	LastUpdated           time.Time           `json:"last_updated"`
	UptimeHistory90Days   []string            `json:"uptime_history_90_days"`
	UptimeSummary         UptimeSummary       `json:"uptime_summary"`
	ResponseTimeSummary7D ResponseTimeSummary `json:"response_time_summary_7_days"`
	RecentEvents          []ResourceEvent     `json:"recent_events"`
	Maintenance           *MaintenanceBanner  `json:"maintenance,omitempty"`
}

// MaintenanceBanner contains minimal info to display a maintenance notice
type MaintenanceBanner struct {
	Status   string     `json:"status"` // scheduled | active
	Title    string     `json:"title"`
	StartAt  *time.Time `json:"start_at,omitempty"`
	EndAt    *time.Time `json:"end_at,omitempty"`
	Timezone *string    `json:"timezone,omitempty"`
}

// UptimeSummary represents uptime percentages for different time windows
type UptimeSummary struct {
	Last24Hours float64 `json:"last_24_hours"`
	Last7Days   float64 `json:"last_7_days"`
	Last30Days  float64 `json:"last_30_days"`
	Last90Days  float64 `json:"last_90_days"`
}

// ResponseTimeSummary represents response time statistics
type ResponseTimeSummary struct {
	AvgMs int `json:"avg_ms"`
	MinMs int `json:"min_ms"`
	MaxMs int `json:"max_ms"`
}

// ResourceEvent represents a status change event (up/down)
type ResourceEvent struct {
	Type      string    `json:"type"` // "up" or "down"
	Timestamp time.Time `json:"timestamp"`
	Duration  *string   `json:"duration,omitempty"`
	Reason    string    `json:"reason"`
	Details   *string   `json:"details,omitempty"`
}
