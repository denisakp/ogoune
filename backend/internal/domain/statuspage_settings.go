package domain

// StatusPageSettings represents the configuration for the public status page
type StatusPageSettings struct {
	Base
	Name                 string `json:"name" gorm:"default:'Status Page'"`
	HomepageURL          string `json:"homepage_url"`
	CustomDomain         string `json:"custom_domain"`
	GoogleAnalyticsID    string `json:"google_analytics_id"`
	EnableDetailsPage    bool   `json:"enable_details_page" gorm:"default:true"`
	ShowUptimePercentage bool   `json:"show_uptime_percentage" gorm:"default:true"`
	HidePausedMonitors   bool   `json:"hide_paused_monitors" gorm:"default:true"`
	ShowIncidentHistory  bool   `json:"show_incident_history" gorm:"default:true"`
}

func (StatusPageSettings) TableName() string { return "status_page_settings" }
