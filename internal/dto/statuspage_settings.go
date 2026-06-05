package dto

import "github.com/denisakp/ogoune/internal/domain"

// StatusPageSettingsRequest represents the request body for updating status page settings.
type StatusPageSettingsRequest struct {
	Name                 string            `json:"name"`
	HomepageURL          string            `json:"homepage_url"`
	CustomDomain         string            `json:"custom_domain"`
	GoogleAnalyticsID    string            `json:"google_analytics_id"`
	EnableDetailsPage    bool              `json:"enable_details_page"`
	ShowUptimePercentage bool              `json:"show_uptime_percentage"`
	HidePausedMonitors   bool              `json:"hide_paused_monitors"`
	ShowIncidentHistory  bool              `json:"show_incident_history"`
	// Spec 060 / US5 — branding
	LogoURLLight   string            `json:"logo_url_light,omitempty"`
	LogoURLDark    string            `json:"logo_url_dark,omitempty"`
	FaviconURL     string            `json:"favicon_url,omitempty"`
	PrimaryColor   string            `json:"primary_color,omitempty"`
	ThemeOverrides map[string]string `json:"theme_overrides,omitempty"`
}

// StatusPageSettingsResponse represents the response body for status page settings.
// Spec 059 fold: domain DNS state now lives on the same row.
type StatusPageSettingsResponse struct {
	ID                       string             `json:"id"`
	Name                     string             `json:"name"`
	HomepageURL              string             `json:"homepage_url"`
	CustomDomain             string             `json:"custom_domain"`
	GoogleAnalyticsID        string             `json:"google_analytics_id"`
	EnableDetailsPage        bool               `json:"enable_details_page"`
	ShowUptimePercentage     bool               `json:"show_uptime_percentage"`
	HidePausedMonitors       bool               `json:"hide_paused_monitors"`
	ShowIncidentHistory      bool               `json:"show_incident_history"`
	CustomDomainStatus       string             `json:"custom_domain_status"`
	CustomDomainSSLStatus    string             `json:"custom_domain_ssl_status"`
	CustomDomainDNSRecords []domain.DNSRecord `json:"custom_domain_dns_records"`
	LogoURLLight           string             `json:"logo_url_light"`
	LogoURLDark            string             `json:"logo_url_dark"`
	FaviconURL             string             `json:"favicon_url"`
	PrimaryColor           string             `json:"primary_color"`
	ThemeOverrides         map[string]string  `json:"theme_overrides"`
	CreatedAt              string             `json:"created_at"`
	UpdatedAt              string             `json:"updated_at"`
}
