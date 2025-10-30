package dto

// StatsSummaryResponse represents aggregated uptime and incident statistics
// for all monitored resources within a given time range.
type StatsSummaryResponse struct {
	Range                    string  `json:"range"`                      // Time range: 2h, 24h, 7d, 30d
	OverallUptime            float64 `json:"overall_uptime"`             // Average uptime percentage across all resources
	Incidents                int     `json:"incidents"`                  // Total number of incidents in the time range
	WithoutIncidentsDuration string  `json:"without_incidents_duration"` // Duration without incidents (e.g., "2h 30m")
	AffectedMonitors         int     `json:"affected_monitors"`          // Number of distinct resources with incidents
}
