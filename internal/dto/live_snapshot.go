package dto

import (
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

// LiveStats contains uptime percentages and response time metrics for a resource.
// All fields are nullable: nil means no data is available for that window.
type LiveStats struct {
	Uptime2h           *float64 `json:"uptime_2h"`
	Uptime24h          *float64 `json:"uptime_24h"`
	Uptime7d           *float64 `json:"uptime_7d"`
	Uptime30d          *float64 `json:"uptime_30d"`
	AvgResponseTime24h *int     `json:"avg_response_time_24h"`
	LastResponseTime   *int     `json:"last_response_time"`
}

// LiveActiveIncident represents a currently unresolved incident for a resource.
type LiveActiveIncident struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	Cause     string    `json:"cause"`
}

// LiveSnapshotResponse is the payload returned by GET /resources/:id/live.
// Stats is a value type (never nil); individual Stats fields may be nil when no data exists.
// ActiveIncident is nil when no unresolved incident exists.
type LiveSnapshotResponse struct {
	Resource         *domain.Resource             `json:"resource"`
	Stats            LiveStats                    `json:"stats"`
	ActiveIncident   *LiveActiveIncident          `json:"active_incident"`
	RecentActivities []*domain.MonitoringActivity `json:"recent_activities"`
	FetchedAt        time.Time                    `json:"fetched_at"`
}
