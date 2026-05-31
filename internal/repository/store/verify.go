package store

import "github.com/denisakp/ogoune/internal/port"

// Compile-time interface satisfaction checks.
var (
	_ port.TagsRepository               = (*TagsRepositoryImpl)(nil)
	_ port.TagsRepository               = (*TagsRepositorySQLC)(nil)
	_ port.ResourceRepository            = (*ResourceRepositoryImpl)(nil)
	_ port.ComponentRepository           = (*ComponentRepositoryImpl)(nil)
	_ port.IncidentRepository            = (*IncidentRepositoryImpl)(nil)
	_ port.IncidentEventStepRepository   = (*IncidentEventStepRepositoryImpl)(nil)
	_ port.NotificationRepository        = (*NotificationRepositoryImpl)(nil)
	_ port.MonitoringActivityRepository  = (*MonitoringActivityRepositoryImpl)(nil)
	_ port.NotificationChannelRepository = (*NotificationChannelRepository)(nil)
	_ port.NotificationChannelRepository = (*NotificationChannelRepositorySQLC)(nil)
	_ port.MaintenanceRepository         = (*MaintenanceRepository)(nil)
	_ port.StatusPageSettingsRepository  = (*StatusPageSettingsRepository)(nil)
	_ port.StatusPageSettingsRepository  = (*StatusPageSettingsRepositorySQLC)(nil)
	_ port.UserRepository                = (*UserRepository)(nil)
	_ port.UserRepository                = (*UserRepositorySQLC)(nil)
	_ port.APIKeyRepository              = (*APIKeyRepositoryImpl)(nil)
	_ port.APIKeyRepository              = (*APIKeyRepositorySQLC)(nil)
	_ port.IncidentDiagnosticsRepository = (*IncidentDiagnosticsRepositoryImpl)(nil)
	_ port.IncidentDiagnosticsRepository = (*IncidentDiagnosticsRepositorySQLC)(nil)
	_ port.ExpiryNotificationLogRepository = (*ExpiryNotificationLogRepository)(nil)
	_ port.ExpiryNotificationLogRepository = (*ExpiryNotificationLogRepositorySQLC)(nil)
	_ port.ResourceCredentialRepository    = (*ResourceCredentialRepository)(nil)
	_ port.ResourceCredentialRepository    = (*ResourceCredentialRepositorySQLC)(nil)
)
