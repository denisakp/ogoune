package dto

// StatusPageSettingsRequest represents the request body for updating status page settings
type StatusPageSettingsRequest struct {
	Name                 string `json:"name"`
	HomepageURL          string `json:"homepage_url"`
	CustomDomain         string `json:"custom_domain"`
	GoogleAnalyticsID    string `json:"google_analytics_id"`
	EnableDetailsPage    bool   `json:"enable_details_page"`
	ShowUptimePercentage bool   `json:"show_uptime_percentage"`
	HidePausedMonitors   bool   `json:"hide_paused_monitors"`
	ShowIncidentHistory  bool   `json:"show_incident_history"`
}

// StatusPageSettingsResponse represents the response body for status page settings
type StatusPageSettingsResponse struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	HomepageURL          string `json:"homepage_url"`
	CustomDomain         string `json:"custom_domain"`
	GoogleAnalyticsID    string `json:"google_analytics_id"`
	EnableDetailsPage    bool   `json:"enable_details_page"`
	ShowUptimePercentage bool   `json:"show_uptime_percentage"`
	HidePausedMonitors   bool   `json:"hide_paused_monitors"`
	ShowIncidentHistory  bool   `json:"show_incident_history"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}
