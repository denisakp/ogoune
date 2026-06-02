package store

import "github.com/denisakp/ogoune/internal/port"

// Compile-time guarantees that every sqlc-backed wrapper satisfies its port
// interface. The legacy GORM compile-time checks were removed in spec 052.
var (
	_ port.TagsRepository                  = (*TagsRepositorySQLC)(nil)
	_ port.ResourceRepository              = (*ResourceRepositorySQLC)(nil)
	_ port.ComponentRepository             = (*ComponentRepositorySQLC)(nil)
	_ port.IncidentRepository              = (*IncidentRepositorySQLC)(nil)
	_ port.IncidentEventStepRepository     = (*IncidentEventStepRepositorySQLC)(nil)
	_ port.NotificationRepository          = (*NotificationRepositorySQLC)(nil)
	_ port.MonitoringActivityRepository    = (*MonitoringActivityRepositorySQLC)(nil)
	_ port.NotificationChannelRepository   = (*NotificationChannelRepositorySQLC)(nil)
	_ port.MaintenanceRepository           = (*MaintenanceRepositorySQLC)(nil)
	_ port.StatusPageSettingsRepository    = (*StatusPageSettingsRepositorySQLC)(nil)
	_ port.UserRepository                  = (*UserRepositorySQLC)(nil)
	_ port.APIKeyRepository                = (*APIKeyRepositorySQLC)(nil)
	_ port.IncidentDiagnosticsRepository   = (*IncidentDiagnosticsRepositorySQLC)(nil)
	_ port.ExpiryNotificationLogRepository = (*ExpiryNotificationLogRepositorySQLC)(nil)
	_ port.ResourceCredentialRepository    = (*ResourceCredentialRepositorySQLC)(nil)
)
