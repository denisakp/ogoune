CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    color TEXT,
    description TEXT
);

CREATE TABLE IF NOT EXISTS components (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    last_notification_status TEXT NOT NULL DEFAULT 'up'
);

CREATE TABLE IF NOT EXISTS resources (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    interval INTEGER NOT NULL DEFAULT 300,
    timeout INTEGER NOT NULL DEFAULT 10,
    target TEXT NOT NULL,
    last_checked DATETIME,
    status TEXT NOT NULL DEFAULT 'pending',
    is_active INTEGER NOT NULL DEFAULT 1,
    failure_count INTEGER NOT NULL DEFAULT 0,
    ssl_expiration_date DATETIME,
    ssl_issuer TEXT,
    domain_expiration_date DATETIME,
    domain_registrar TEXT,
    component_id TEXT,
    FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS incidents (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    resource_id TEXT NOT NULL,
    cause TEXT NOT NULL DEFAULT 'unknown_failure',
    resolved_at DATETIME,
    started_at DATETIME NOT NULL,
    details BLOB,
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS incident_event_steps (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    incident_id TEXT NOT NULL,
    step TEXT NOT NULL,
    message TEXT,
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS incident_diagnostics (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
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
    body_truncated INTEGER NOT NULL DEFAULT 0,
    body_encoded INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification_events (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    incident_id TEXT NOT NULL,
    type TEXT NOT NULL,
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification_channels (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config BLOB NOT NULL,
    enabled_by_default INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS maintenances (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    strategy TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT '',
    start_at DATETIME,
    end_at DATETIME,
    cron_expr TEXT,
    window_minutes INTEGER,
    timezone TEXT,
    effective_from DATETIME,
    effective_until DATETIME,
    started_at DATETIME,
    ended_at DATETIME
);

CREATE TABLE IF NOT EXISTS monitoring_activities (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    resource_id TEXT NOT NULL,
    message TEXT NOT NULL,
    success INTEGER NOT NULL DEFAULT 0,
    response_time INTEGER NOT NULL DEFAULT 0,
    response_data BLOB,
    is_maintenance INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS status_page_settings (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL DEFAULT 'Status Page',
    homepage_url TEXT NOT NULL DEFAULT '',
    custom_domain TEXT NOT NULL DEFAULT '',
    google_analytics_id TEXT NOT NULL DEFAULT '',
    enable_details_page INTEGER NOT NULL DEFAULT 1,
    show_uptime_percentage INTEGER NOT NULL DEFAULT 1,
    hide_paused_monitors INTEGER NOT NULL DEFAULT 1,
    show_incident_history INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    hashed_password TEXT NOT NULL,
    password_initialized INTEGER NOT NULL DEFAULT 0,
    force_password_change INTEGER NOT NULL DEFAULT 0,
    two_factor_enabled INTEGER NOT NULL DEFAULT 0,
    two_factor_secret TEXT NOT NULL DEFAULT '',
    two_factor_backup_codes BLOB,
    last_login_at DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS resource_tags (
    resource_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (resource_id, tag_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS resource_notification_channels (
    resource_id TEXT NOT NULL,
    notification_channel_id TEXT NOT NULL,
    PRIMARY KEY (resource_id, notification_channel_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS component_notification_channels (
    component_id TEXT NOT NULL,
    notification_channel_id TEXT NOT NULL,
    PRIMARY KEY (component_id, notification_channel_id),
    FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS maintenance_resources (
    maintenance_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    PRIMARY KEY (maintenance_id, resource_id),
    FOREIGN KEY (maintenance_id) REFERENCES maintenances(id) ON DELETE CASCADE,
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);