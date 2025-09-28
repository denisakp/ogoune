package domain

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// / define a base model with common fields
type Base struct {
	ID        string    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
}

// / hook to set timestamps before creating a record
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
	Name      string      `gorm:"uniqueIndex"`
	Resources []*Resource `gorm:"many2many:resource_tags;"`
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
	Name         string
	Type         ResourceType `gorm:"index"`
	Interval     int
	Timeout      int
	Target       string
	LastChecked  *time.Time
	Status       ResourceStatus `gorm:"default:pending"`
	IsActive     bool           `gorm:"default:true"`
	FailureCount int            `gorm:"default:0"`
	Incidents    []Incident     `json:"-"`
	Tags         []*Tags        `gorm:"many2many:resource_tags;"`
}

func (Resource) TableName() string { return "resources" }

// An Incident represents an event where a Resource is down or experiencing issues.
type Incident struct {
	Base
	ResourceID string   `gorm:"index"`
	Resource   Resource `gorm:"foreignKey:ResourceID"`
	Reason     string
	IsResolved bool       `gorm:"default:false"`
	ResolvedAt *time.Time `gorm:"index"`
	StartedAt  time.Time
	Details    []byte
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
	IncidentID string                `gorm:"index"`
	Incident   Incident              `gorm:"foreignKey:IncidentID"`
	Step       IncidentEventStepType `gorm:"index"`
	Message    *string
}

func (IncidentEventStep) TableName() string { return "incident_event_steps" }

type IntegrationType string

const (
	IntegrationSMTP       IntegrationType = "smtp"
	IntegrationSlack      IntegrationType = "slack"
	IntegrationGoogleChat IntegrationType = "google_chat"
)

type Integration struct {
	Base
	Name     string
	Target   string
	Type     IntegrationType `gorm:"index"`
	IsActive bool            `gorm:"default:true"`
}

func (Integration) TableName() string { return "integrations" }

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
	IncidentID string                `gorm:"index"`
	Incident   Incident              `gorm:"foreignKey:IncidentID"`
	Type       NotificationEventType `gorm:"index"`
}
