package dto

import (
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

type CreateResourcePayload struct {
	Name     string              `json:"name" binding:"required"`
	Type     domain.ResourceType `json:"type" binding:"required"`
	Interval int                 `json:"interval" binding:"required,min=10,max=3600"`
	Timeout  int                 `json:"timeout" binding:"required,min=1,max=60"`
	Target   string              `json:"target" binding:"required,url"`
	Tags     []string            `json:"tags"`
}

// UpdateResourcePayload contains the fields that can be updated for a resource
type UpdateResourcePayload struct {
	Name     *string              `json:"name,omitempty"`
	Type     *domain.ResourceType `json:"type,omitempty"`
	Target   *string              `json:"target,omitempty"`
	Interval *int                 `json:"interval,omitempty"`
	Timeout  *int                 `json:"timeout,omitempty"`
	IsActive *bool                `json:"is_active,omitempty"`
	Tags     *[]string            `json:"tags,omitempty"`
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

// ResourceResponse represents the enriched resource response with response times
type ResourceResponse struct {
	domain.Resource
	ResponseTimes []ResponseTimePoint `json:"response_times,omitempty"`
}
