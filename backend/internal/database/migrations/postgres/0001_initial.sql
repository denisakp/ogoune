CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    name TEXT NOT NULL,
    color TEXT,
    description TEXT
);

CREATE TABLE IF NOT EXISTS components (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    last_notification_status TEXT NOT NULL DEFAULT 'up'
);

CREATE TABLE IF NOT EXISTS resources (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    interval INTEGER NOT NULL DEFAULT 300,
    timeout INTEGER NOT NULL DEFAULT 10,
    target TEXT NOT NULL,
    last_checked TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'pending',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    failure_count INTEGER NOT NULL DEFAULT 0,
    ssl_expiration_date TIMESTAMPTZ,
    ssl_issuer TEXT,
    domain_expiration_date TIMESTAMPTZ,
    domain_registrar TEXT,
    component_id TEXT,
    CONSTRAINT fk_resources_component FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS incidents (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    resource_id TEXT NOT NULL,
    cause TEXT NOT NULL DEFAULT 'unknown_failure',
    resolved_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ NOT NULL,
    details BYTEA,
    CONSTRAINT fk_incidents_resource FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS incident_event_steps (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    incident_id TEXT NOT NULL,
    step TEXT NOT NULL,
    message TEXT,
    CONSTRAINT fk_incident_event_steps_incident FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS incident_diagnostics (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    incident_id TEXT NOT NULL,
    request_method TEXT NOT NULL DEFAULT '',
    request_url TEXT NOT NULL DEFAULT '',
    request_headers TEXT NOT NULL DEFAULT '{}',
    request_timeout INTEGER NOT NULL DEFAULT 0,
    http_status_code INTEGER NOT NULL DEFAULT -1,
    response_headers TEXT NOT NULL DEFAULT '{}',
    response_body TEXT NOT NULL DEFAULT '',
    response_size INTEGER NOT NULL DEFAULT 0,
    failure_type TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    error_summary TEXT NOT NULL DEFAULT '',
    total_duration INTEGER NOT NULL DEFAULT 0,
    dns_duration INTEGER NOT NULL DEFAULT 0,
    tls_duration INTEGER NOT NULL DEFAULT 0,
    first_byte_duration INTEGER NOT NULL DEFAULT 0,
    body_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    body_encoded BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_incident_diagnostics_incident FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification_events (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    incident_id TEXT NOT NULL,
    type TEXT NOT NULL,
    CONSTRAINT fk_notification_events_incident FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification_channels (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config BYTEA NOT NULL,
    enabled_by_default BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS maintenances (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    strategy TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT '',
    start_at TIMESTAMPTZ,
    end_at TIMESTAMPTZ,
    cron_expr TEXT,
    window_minutes INTEGER,
    timezone TEXT,
    effective_from TIMESTAMPTZ,
    effective_until TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS monitoring_activities (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    resource_id TEXT NOT NULL,
    message TEXT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    response_time INTEGER NOT NULL DEFAULT 0,
    response_data BYTEA,
    is_maintenance BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_monitoring_activities_resource FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS status_page_settings (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    name TEXT NOT NULL DEFAULT 'Status Page',
    homepage_url TEXT NOT NULL DEFAULT '',
    custom_domain TEXT NOT NULL DEFAULT '',
    google_analytics_id TEXT NOT NULL DEFAULT '',
    enable_details_page BOOLEAN NOT NULL DEFAULT TRUE,
    show_uptime_percentage BOOLEAN NOT NULL DEFAULT TRUE,
    hide_paused_monitors BOOLEAN NOT NULL DEFAULT TRUE,
    show_incident_history BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    hashed_password TEXT NOT NULL,
    password_initialized BOOLEAN NOT NULL DEFAULT FALSE,
    force_password_change BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_secret TEXT NOT NULL DEFAULT '',
    two_factor_backup_codes BYTEA,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS resource_tags (
    resource_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (resource_id, tag_id),
    CONSTRAINT fk_resource_tags_resource FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    CONSTRAINT fk_resource_tags_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS resource_notification_channels (
    resource_id TEXT NOT NULL,
    notification_channel_id TEXT NOT NULL,
    PRIMARY KEY (resource_id, notification_channel_id),
    CONSTRAINT fk_resource_notification_channels_resource FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    CONSTRAINT fk_resource_notification_channels_channel FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS component_notification_channels (
    component_id TEXT NOT NULL,
    notification_channel_id TEXT NOT NULL,
    PRIMARY KEY (component_id, notification_channel_id),
    CONSTRAINT fk_component_notification_channels_component FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE CASCADE,
    CONSTRAINT fk_component_notification_channels_channel FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS maintenance_resources (
    maintenance_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    PRIMARY KEY (maintenance_id, resource_id),
    CONSTRAINT fk_maintenance_resources_maintenance FOREIGN KEY (maintenance_id) REFERENCES maintenances(id) ON DELETE CASCADE,
    CONSTRAINT fk_maintenance_resources_resource FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);