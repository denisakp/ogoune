package dto

import "time"

// MaintenanceCreatePayload defines the request body for creating a maintenance window
type MaintenanceCreatePayload struct {
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	Strategy       string     `json:"strategy"`
	StartAt        *time.Time `json:"start_at,omitempty"`
	EndAt          *time.Time `json:"end_at,omitempty"`
	CronExpr       *string    `json:"cron_expr,omitempty"`
	WindowMinutes  *int       `json:"window_minutes,omitempty"`
	Timezone       *string    `json:"timezone,omitempty"`
	EffectiveFrom  *time.Time `json:"effective_from,omitempty"`
	EffectiveUntil *time.Time `json:"effective_until,omitempty"`
	ResourceIDs    []string   `json:"resource_ids"`
}

// MaintenanceUpdatePayload defines the request body for updating a maintenance window
type MaintenanceUpdatePayload struct {
	Title          *string    `json:"title,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Strategy       *string    `json:"strategy,omitempty"`
	StartAt        *time.Time `json:"start_at,omitempty"`
	EndAt          *time.Time `json:"end_at,omitempty"`
	CronExpr       *string    `json:"cron_expr,omitempty"`
	WindowMinutes  *int       `json:"window_minutes,omitempty"`
	Timezone       *string    `json:"timezone,omitempty"`
	EffectiveFrom  *time.Time `json:"effective_from,omitempty"`
	EffectiveUntil *time.Time `json:"effective_until,omitempty"`
	ResourceIDs    []string   `json:"resource_ids,omitempty"`
}

// MaintenanceResponse is the response representation of a maintenance window
type MaintenanceResponse struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	Strategy       string     `json:"strategy"`
	Status         string     `json:"status"`
	StartAt        *time.Time `json:"start_at,omitempty"`
	EndAt          *time.Time `json:"end_at,omitempty"`
	CronExpr       *string    `json:"cron_expr,omitempty"`
	WindowMinutes  *int       `json:"window_minutes,omitempty"`
	Timezone       *string    `json:"timezone,omitempty"`
	EffectiveFrom  *time.Time `json:"effective_from,omitempty"`
	EffectiveUntil *time.Time `json:"effective_until,omitempty"`
	ResourceIDs    []string   `json:"resource_ids"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}
