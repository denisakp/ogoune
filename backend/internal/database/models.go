package database

import domain "github.com/denisakp/pulseguard/internal/domain"

// RegisteredModels provides a single registry for schema-validation tests.
var RegisteredModels = []any{
	&domain.Component{},
	&domain.Resource{},
	&domain.Incident{},
	&domain.IncidentEventStep{},
	&domain.IncidentDiagnostics{},
	&domain.NotificationEvent{},
	&domain.NotificationChannel{},
	&domain.Maintenance{},
	&domain.Tags{},
	&domain.MonitoringActivity{},
	&domain.StatusPageSettings{},
	&domain.User{},
}

// RegisteredJoinTables captures SQL-managed many-to-many tables that are not first-class domain models.
var RegisteredJoinTables = []string{
	"resource_tags",
	"resource_notification_channels",
	"component_notification_channels",
	"maintenance_resources",
}
