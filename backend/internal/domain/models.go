package domain

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// Base define a base model with common fields
type Base struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook to set timestamps before creating a record
func (base *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == "" {
		// generate a new ULID for the ID field
		t := time.Now()
		entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		base.ID = ulid.MustNew(ulid.Timestamp(t), entropy).String()
	}
	return
}

// Tags represents a tag that can be associated with multiple resources
type Tags struct {
	Base
	Name        string      `json:"name" gorm:"uniqueIndex"`
	Color       *string     `json:"color,omitempty"`
	Description *string     `json:"description,omitempty"`
	Resources   []*Resource `json:"resources" gorm:"many2many:resource_tags;"`
}

func (Tags) TableName() string { return "tags" }

type ResourceType string

const (
	ResourceHTTP ResourceType = "http"
	ResourceTCP  ResourceType = "tcp"
)

type ResourceStatus string

const (
	StatusUp      ResourceStatus = "up"
	StatusDown    ResourceStatus = "down"
	StatusError   ResourceStatus = "error"
	StatusUnknown ResourceStatus = "unknown"
	StatusPaused  ResourceStatus = "paused"
	StatusPending ResourceStatus = "pending"
	StatusWarn    ResourceStatus = "warning"
)

// ResourceMetaData collect domain and ssl metadata form resource
type ResourceMetaData struct {
	SSLExpirationDate    *time.Time `json:"ssl_expiration_date" gorm:"column:ssl_expiration_date"`
	SSLIssuer            string     `json:"ssl_issuer" gorm:"column:ssl_issuer"`
	DomainExpirationDate *time.Time `json:"domain_expiration_date" gorm:"column:domain_expiration_date"`
	DomainRegistrar      string     `json:"domain_registrar" gorm:"column:domain_registrar"`
}

// A Resource is something that can be monitored, such as a website or server.
type Resource struct {
	Base
	Name         string            `json:"name"`
	Type         ResourceType      `json:"type"  gorm:"index"`
	Interval     int               `json:"interval" gorm:"default:300"` // in seconds
	Timeout      int               `json:"timeout" gorm:"default:10"`   // in seconds
	Target       string            `json:"target"`
	LastChecked  *time.Time        `json:"last_checked"`
	Status       ResourceStatus    `json:"status" gorm:"default:pending"`
	IsActive     bool              `json:"is_active" gorm:"default:true"`
	FailureCount int               `json:"failure_count" gorm:"default:0"`
	Metadata     *ResourceMetaData `json:"metadata" gorm:"embedded"`
	Incidents    []Incident        `json:"incidents"`
	Tags         []*Tags           `json:"tags" gorm:"many2many:resource_tags;"`
}

func (Resource) TableName() string { return "resources" }

// Incident represents an event where a Resource is down or experiencing issues.
type Incident struct {
	Base
	ResourceID string              `json:"resource_id" gorm:"index"`
	Resource   Resource            `json:"resource" gorm:"foreignKey:ResourceID"`
	Cause      string              `json:"cause" gorm:"index;default:unknown_failure"`
	ResolvedAt *time.Time          `json:"resolved_at" gorm:"index"` // nil = active, timestamp = resolved
	StartedAt  time.Time           `json:"started_at" gorm:"index"`
	Details    []byte              `json:"details"`
	EventStep  []IncidentEventStep `json:"event_steps"`
}

func (Incident) TableName() string { return "incidents" }

type IncidentEventStepType string

const (
	IncidentEventStepDetected  IncidentEventStepType = "detected"
	IncidentEventStepResolved  IncidentEventStepType = "resolved"
	IncidentEventStepAlert     IncidentEventStepType = "alert_sent"
	IncidentEventStepDownAlert IncidentEventStepType = "resource_down_alert"
	IncidentEventStepUpAlert   IncidentEventStepType = "resource_up_alert"
)

// IncidentEventStep represents a step in the lifecycle of an incident, such as detection or resolution.
type IncidentEventStep struct {
	Base
	IncidentID string                `json:"incident_id" gorm:"index"`
	Incident   Incident              `json:"incident" gorm:"foreignKey:IncidentID"`
	Step       IncidentEventStepType `json:"step" gorm:"index"`
	Message    *string               `json:"message"`
}

func (IncidentEventStep) TableName() string { return "incident_event_steps" }

// EventType Event type constants for notification event types (avoid magic strings)
type EventType string

const (
	EventTypeDown   EventType = "down"
	EventTypeUp     EventType = "up"
	EventTypeExpiry EventType = "expiry"
)

type NotificationEventType string

const (
	NotificationEventTypeUp     NotificationEventType = "up"
	NotificationEventTypeDown   NotificationEventType = "down"
	NotificationEventTypeExpiry NotificationEventType = "expiry"
)

type NotificationEventStatusType string

const (
	NotificationEventStatusSent      NotificationEventStatusType = "sent"
	NotificationEventStatusFailed    NotificationEventStatusType = "failed"
	NotificationEventStatusPending   NotificationEventStatusType = "pending"
	NotificationEventStatusDelivered NotificationEventStatusType = "delivered"
	NotificationEventStatusRead      NotificationEventStatusType = "read"
)

type NotificationEvent struct {
	Base
	IncidentID string                `json:"incident_id" gorm:"index"`
	Incident   Incident              `json:"incident" gorm:"foreignKey:IncidentID"`
	Type       NotificationEventType `json:"type" gorm:"index"`
}

func (NotificationEvent) TableName() string { return "notification_events" }

type MonitoringActivity struct {
	Base
	ResourceID   string   `json:"resource_id" gorm:"index"`
	Resource     Resource `json:"resource" gorm:"foreignKey:ResourceID"`
	Message      string   `json:"message"`
	Success      bool     `json:"success"`
	ResponseTime int      `json:"response_time"`
	ResponseData []byte   `json:"response_data"`
}

func (MonitoringActivity) TableName() string { return "monitoring_activities" }

// UptimeStat represents aggregated uptime data for a specific hour
type UptimeStat struct {
	Hour            time.Time `json:"hour"`
	UptimePercent   float64   `json:"uptime_percent"`
	SuccessfulCount int       `json:"successful_count"`
	TotalCount      int       `json:"total_count"`
}

// ResponseTimePoint represents a single response time measurement with timestamp
type ResponseTimePoint struct {
	Timestamp    time.Time `json:"timestamp"`
	ResponseTime int       `json:"response_time"` // in milliseconds
}

// GlobalStats represents aggregated statistics across all monitored resources
type GlobalStats struct {
	OverallUptime            float64 // Average uptime percentage across all resources
	TotalIncidents           int     // Total number of incidents in the time range
	WithoutIncidentsDuration int64   // Duration in seconds without any incidents
	AffectedMonitors         int     // Number of distinct resources with incidents
}
