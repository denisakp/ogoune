package domain

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/datatypes"
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
)

// A Resource is something that can be monitored, such as a website or server.
type Resource struct {
	Base
	Name         string         `json:"name"`
	Type         ResourceType   `json:"type"  gorm:"index"`
	Interval     int            `json:"interval" gorm:"default:300"` // in seconds
	Timeout      int            `json:"timeout" gorm:"default:10"`   // in seconds
	Target       string         `json:"target"`
	LastChecked  *time.Time     `json:"last_checked"`
	Status       ResourceStatus `json:"status" gorm:"default:pending"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	FailureCount int            `json:"failure_count" gorm:"default:0"`
	Incidents    []Incident     `json:"-"`
	Tags         []*Tags        `json:"tags" gorm:"many2many:resource_tags;"`
}

func (Resource) TableName() string { return "resources" }

// Incident represents an event where a Resource is down or experiencing issues.
type Incident struct {
	Base
	ResourceID string              `json:"resource_id" gorm:"index"`
	Resource   Resource            `json:"resource" gorm:"foreignKey:ResourceID"`
	Reason     string              `json:"reason"`
	Cause      string              `json:"cause" gorm:"index;default:unknown_failure"`
	ResolvedAt *time.Time          `json:"resolved_at" gorm:"index"` // nil = active, timestamp = resolved
	StartedAt  time.Time           `json:"started_at" gorm:"index"`
	Details    []byte              `json:"details"`
	EventStep  []IncidentEventStep `json:"-"`
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

type IntegrationType string

const (
	IntegrationSlack      IntegrationType = "slack"
	IntegrationGoogleChat IntegrationType = "google_chat"
	IntegrationDiscord    IntegrationType = "discord"
)

// EventType Event type constants for notification event types (avoid magic strings)
type EventType string

const (
	EventTypeDown   EventType = "down"
	EventTypeUp     EventType = "up"
	EventTypeExpiry EventType = "expiry"
)

type Integration struct {
	Base
	Name       string         `json:"name"`
	IsActive   bool           `json:"is_active" gorm:"default:true"`
	Config     datatypes.JSON `json:"config"`      // Stores integration-specific config, e.g., {"type": "slack", "webhook_url": "..."}
	EventTypes datatypes.JSON `json:"event_types"` // Stores []string, e.g., ["down", "up"]
}

func (Integration) TableName() string { return "integrations" }

// GetType extracts the integration type from the Config JSON field
func (i *Integration) GetType() IntegrationType {
	var config map[string]interface{}
	if err := json.Unmarshal(i.Config, &config); err != nil {
		return ""
	}
	if typeVal, ok := config["type"].(string); ok {
		return IntegrationType(typeVal)
	}
	return ""
}

// GetConfig returns the full config as a map
func (i *Integration) GetConfig() (map[string]interface{}, error) {
	var config map[string]interface{}
	if err := json.Unmarshal(i.Config, &config); err != nil {
		return nil, err
	}
	return config, nil
}

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
