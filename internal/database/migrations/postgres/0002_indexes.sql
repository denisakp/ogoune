CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_created_at ON tags(created_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_components_name ON components(name);
CREATE INDEX IF NOT EXISTS idx_components_created_at ON components(created_at);

CREATE INDEX IF NOT EXISTS idx_resources_created_at ON resources(created_at);
CREATE INDEX IF NOT EXISTS idx_resources_type ON resources(type);
CREATE INDEX IF NOT EXISTS idx_resources_status ON resources(status);
CREATE INDEX IF NOT EXISTS idx_resources_component_id ON resources(component_id);

CREATE INDEX IF NOT EXISTS idx_incidents_created_at ON incidents(created_at);
CREATE INDEX IF NOT EXISTS idx_incidents_resource_id ON incidents(resource_id);
CREATE INDEX IF NOT EXISTS idx_incidents_cause ON incidents(cause);
CREATE INDEX IF NOT EXISTS idx_incidents_resolved_at ON incidents(resolved_at);
CREATE INDEX IF NOT EXISTS idx_incidents_started_at ON incidents(started_at);

CREATE INDEX IF NOT EXISTS idx_incident_event_steps_created_at ON incident_event_steps(created_at);
CREATE INDEX IF NOT EXISTS idx_incident_event_steps_incident_id ON incident_event_steps(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_event_steps_step ON incident_event_steps(step);

CREATE INDEX IF NOT EXISTS idx_incident_diagnostics_created_at ON incident_diagnostics(created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_incident_diagnostics_incident_id ON incident_diagnostics(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_diagnostics_http_status_code ON incident_diagnostics(http_status_code);
CREATE INDEX IF NOT EXISTS idx_incident_diagnostics_failure_type ON incident_diagnostics(failure_type);

CREATE INDEX IF NOT EXISTS idx_notification_events_created_at ON notification_events(created_at);
CREATE INDEX IF NOT EXISTS idx_notification_events_incident_id ON notification_events(incident_id);
CREATE INDEX IF NOT EXISTS idx_notification_events_type ON notification_events(type);

CREATE INDEX IF NOT EXISTS idx_notification_channels_created_at ON notification_channels(created_at);
CREATE INDEX IF NOT EXISTS idx_notification_channels_type ON notification_channels(type);

CREATE INDEX IF NOT EXISTS idx_maintenances_created_at ON maintenances(created_at);
CREATE INDEX IF NOT EXISTS idx_maintenances_strategy ON maintenances(strategy);
CREATE INDEX IF NOT EXISTS idx_maintenances_status ON maintenances(status);
CREATE INDEX IF NOT EXISTS idx_maintenances_start_at ON maintenances(start_at);
CREATE INDEX IF NOT EXISTS idx_maintenances_end_at ON maintenances(end_at);
CREATE INDEX IF NOT EXISTS idx_maintenances_cron_expr ON maintenances(cron_expr);
CREATE INDEX IF NOT EXISTS idx_maintenances_effective_from ON maintenances(effective_from);
CREATE INDEX IF NOT EXISTS idx_maintenances_effective_until ON maintenances(effective_until);
CREATE INDEX IF NOT EXISTS idx_maintenances_started_at ON maintenances(started_at);
CREATE INDEX IF NOT EXISTS idx_maintenances_ended_at ON maintenances(ended_at);

CREATE INDEX IF NOT EXISTS idx_monitoring_activities_created_at ON monitoring_activities(created_at);
CREATE INDEX IF NOT EXISTS idx_monitoring_activities_resource_id ON monitoring_activities(resource_id);

CREATE INDEX IF NOT EXISTS idx_status_page_settings_created_at ON status_page_settings(created_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_resource_tags_tag_id ON resource_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_resource_notification_channels_channel_id ON resource_notification_channels(notification_channel_id);
CREATE INDEX IF NOT EXISTS idx_component_notification_channels_channel_id ON component_notification_channels(notification_channel_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_resources_resource_id ON maintenance_resources(resource_id);