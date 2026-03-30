package dto

import (
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// CreateResourcePayload contains fields for creating a new monitoring resource.
// Tags field expects tag names (strings) - tags will be created if they don't exist.
type CreateResourcePayload struct {
	Name                    string              `json:"name" binding:"required"`
	Type                    domain.ResourceType `json:"type" binding:"required"`
	Interval                int                 `json:"interval" binding:"required,min=10,max=3600"`
	Timeout                 int                 `json:"timeout" binding:"required,min=1,max=60"`
	Target                  string              `json:"target" binding:"required"`
	Tags                    []string            `json:"tags"` // Tag names - will be created if they don't exist
	ComponentID             *string             `json:"component_id,omitempty"`
	ConfirmationChecks      *int                `json:"confirmation_checks,omitempty"`
	ConfirmationInterval    *int                `json:"confirmation_interval,omitempty"`
	ExpiryAlertThresholds   *string             `json:"expiry_alert_thresholds,omitempty"`
	FlapDetectionEnabled    *bool               `json:"flap_detection_enabled,omitempty"`
	FlapThreshold           *int                `json:"flap_threshold,omitempty"`
	FlapWindowSeconds       *int                `json:"flap_window_seconds,omitempty"`
	FlapMaxDurationMinutes  *int                `json:"flap_max_duration_minutes,omitempty"`
	ReminderIntervalMinutes *int                `json:"reminder_interval_minutes,omitempty"`
}

// UpdateResourcePayload contains the fields that can be updated for a resource.
// Tags field expects tag IDs (ULIDs) - only existing tags can be associated.
type UpdateResourcePayload struct {
	Name                    *string              `json:"name,omitempty"`
	Type                    *domain.ResourceType `json:"type,omitempty"`
	Target                  *string              `json:"target,omitempty"`
	Interval                *int                 `json:"interval,omitempty"`
	Timeout                 *int                 `json:"timeout,omitempty"`
	IsActive                *bool                `json:"is_active,omitempty"`
	Tags                    *[]string            `json:"tags,omitempty"` // Tag IDs (ULIDs) - must reference existing tags
	ComponentID             *string              `json:"component_id,omitempty"`
	ConfirmationChecks      *int                 `json:"confirmation_checks,omitempty"`
	ConfirmationInterval    *int                 `json:"confirmation_interval,omitempty"`
	ExpiryAlertThresholds   *string              `json:"expiry_alert_thresholds,omitempty"`
	FlapDetectionEnabled    *bool                `json:"flap_detection_enabled,omitempty"`
	FlapThreshold           *int                 `json:"flap_threshold,omitempty"`
	FlapWindowSeconds       *int                 `json:"flap_window_seconds,omitempty"`
	FlapMaxDurationMinutes  *int                 `json:"flap_max_duration_minutes,omitempty"`
	ReminderIntervalMinutes *int                 `json:"reminder_interval_minutes,omitempty"`
}

// UptimeStatResponse represents hourly uptime percentage for the last 24 hours
type UptimeStatResponse struct {
	Hour            time.Time `json:"hour"`
	UptimePercent   float64   `json:"uptime_percent"`
	SuccessfulCount int       `json:"successful_count"`
	TotalCount      int       `json:"total_count"`
}

// ResponseTimePoint represents a single response time measurement
type ResponseTimePoint struct {
	Timestamp    time.Time `json:"timestamp"`
	ResponseTime int       `json:"response_time"` // in milliseconds
}

// ResourceMetaDataResponse extends the domain metadata with computed day-remaining fields.
type ResourceMetaDataResponse struct {
	domain.ResourceMetaData
	SSLDaysRemaining    *int `json:"ssl_days_remaining,omitempty"`
	DomainDaysRemaining *int `json:"domain_days_remaining,omitempty"`
}

// ResourceResponse represents the enriched resource response with response times and computed expiry fields.
type ResourceResponse struct {
	domain.Resource
	ResponseTimes []ResponseTimePoint       `json:"response_times,omitempty"`
	ExpiryStatus  domain.ExpiryStatus       `json:"expiry_status,omitempty"`
	MetadataExt   *ResourceMetaDataResponse `json:"metadata,omitempty"`
}

// enrichMetadata computes the expiry fields and attaches the extended metadata to the response.
// Should be called after the resource's Metadata is populated.
func EnrichResponseExpiry(rr *ResourceResponse) {
	if rr.Resource.Metadata == nil {
		return
	}

	ext := &ResourceMetaDataResponse{ResourceMetaData: *rr.Resource.Metadata}

	var sslDays, domainDays *int

	if rr.Resource.Metadata.SSLExpirationDate != nil {
		d := int(time.Until(*rr.Resource.Metadata.SSLExpirationDate).Hours() / 24)
		sslDays = &d
		ext.SSLDaysRemaining = sslDays
	}
	if rr.Resource.Metadata.DomainExpirationDate != nil {
		d := int(time.Until(*rr.Resource.Metadata.DomainExpirationDate).Hours() / 24)
		domainDays = &d
		ext.DomainDaysRemaining = domainDays
	}

	// Compute SSL and domain expiry statuses
	sslStatus := domain.ExpiryStatusOK
	if sslDays != nil {
		sslStatus = domain.ComputeExpiryStatus(*sslDays)
	}
	domainStatus := domain.ExpiryStatusOK
	if domainDays != nil {
		domainStatus = domain.ComputeExpiryStatus(*domainDays)
	}

	rr.ExpiryStatus = domain.AggregateExpiryStatus(sslStatus, domainStatus)
	rr.MetadataExt = ext
	// Prevent the embedded Metadata field from double-serializing (MetadataExt replaces it).
	rr.Resource.Metadata = nil
}

// ICMPAvailabilityState describes the current ICMP monitoring availability on this host.
type ICMPAvailabilityState struct {
	Enabled             bool   `json:"enabled"`
	CapabilityAvailable bool   `json:"capability_available"`
	Reason              string `json:"reason"`
}

// SystemCapabilitiesResponse is the response body for GET /api/system/capabilities.
type SystemCapabilitiesResponse struct {
	ICMP ICMPAvailabilityState `json:"icmp"`
}
