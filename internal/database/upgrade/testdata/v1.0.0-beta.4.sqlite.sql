PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE schema_migrations (
    version TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME NOT NULL
);
INSERT INTO schema_migrations VALUES('0000','schema_migrations','2026-09-14 10:35:19.69504 +0000 UTC');
INSERT INTO schema_migrations VALUES('0001','initial','2026-09-14 10:35:19.695867 +0000 UTC');
INSERT INTO schema_migrations VALUES('0002','indexes','2026-09-14 10:35:19.697167 +0000 UTC');
INSERT INTO schema_migrations VALUES('0003','confirmation_fields','2026-09-14 10:35:19.697942 +0000 UTC');
INSERT INTO schema_migrations VALUES('0004','expiry_notification_logs','2026-09-14 10:35:19.698408 +0000 UTC');
INSERT INTO schema_migrations VALUES('0005','intelligent_alerting_fields','2026-09-14 10:35:19.700236 +0000 UTC');
INSERT INTO schema_migrations VALUES('0006','live_monitor_indexes','2026-09-14 10:35:19.700394 +0000 UTC');
INSERT INTO schema_migrations VALUES('0007','api_keys','2026-09-14 10:35:19.700607 +0000 UTC');
INSERT INTO schema_migrations VALUES('0008','pending_notification_retry','2026-09-14 10:35:19.702053 +0000 UTC');
INSERT INTO schema_migrations VALUES('0009','icmp_diagnostics_fields','2026-09-14 10:35:19.703122 +0000 UTC');
INSERT INTO schema_migrations VALUES('0010','heartbeat_fields','2026-09-14 10:35:19.70419 +0000 UTC');
INSERT INTO schema_migrations VALUES('0011','keyword_fields','2026-09-14 10:35:19.705411 +0000 UTC');
INSERT INTO schema_migrations VALUES('0012','protocol_fields','2026-09-14 10:35:19.705943 +0000 UTC');
INSERT INTO schema_migrations VALUES('0013','resource_credentials','2026-09-14 10:35:19.706151 +0000 UTC');
INSERT INTO schema_migrations VALUES('0014','sessions.down','2026-09-14 10:35:19.706275 +0000 UTC');
INSERT INTO schema_migrations VALUES('0015','two_factor.down','2026-09-14 10:35:19.706512 +0000 UTC');
INSERT INTO schema_migrations VALUES('0016','escalation_policies.down','2026-09-14 10:35:19.706805 +0000 UTC');
INSERT INTO schema_migrations VALUES('0017','custom_domains.down','2026-09-14 10:35:19.707196 +0000 UTC');
INSERT INTO schema_migrations VALUES('0018','statuspage_domain_state.down','2026-09-14 10:35:19.707447 +0000 UTC');
INSERT INTO schema_migrations VALUES('0019','statuspage_branding.down','2026-09-14 10:35:19.708496 +0000 UTC');
INSERT INTO schema_migrations VALUES('0020','uptime_daily_agg.down','2026-09-14 10:35:19.709933 +0000 UTC');
INSERT INTO schema_migrations VALUES('0021','incident_updates.down','2026-09-14 10:35:19.710193 +0000 UTC');
INSERT INTO schema_migrations VALUES('0022','notification_channel_dispatch_stats.down','2026-09-14 10:35:19.710424 +0000 UTC');
INSERT INTO schema_migrations VALUES('0023','replace_ga_with_umami.down','2026-09-14 10:35:19.711398 +0000 UTC');
INSERT INTO schema_migrations VALUES('0024','notifications.down','2026-09-14 10:35:19.71212 +0000 UTC');
INSERT INTO schema_migrations VALUES('0025','dashboards.down','2026-09-14 10:35:19.712382 +0000 UTC');
INSERT INTO schema_migrations VALUES('0026','report_history.down','2026-09-14 10:35:19.712667 +0000 UTC');
INSERT INTO schema_migrations VALUES('0027','announcements.down','2026-09-14 10:35:19.713079 +0000 UTC');
INSERT INTO schema_migrations VALUES('0028','hosts.down','2026-09-14 10:35:19.713306 +0000 UTC');
INSERT INTO schema_migrations VALUES('0029','host_alert_state.down','2026-09-14 10:35:19.71445 +0000 UTC');
INSERT INTO schema_migrations VALUES('0030','notification_escalation_state.down','2026-09-14 10:35:19.714665 +0000 UTC');
CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    color TEXT,
    description TEXT
);
INSERT INTO tags VALUES('01M2FQRWWR7T15NJMDTRV0XHG7','2026-09-14 10:35:28.536319 +0000 GMT m=+8.850865751','2026-09-14 10:35:28.536319 +0000 GMT m=+8.850865751','b4',NULL,NULL);
CREATE TABLE components (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    last_notification_status TEXT NOT NULL DEFAULT 'up'
, grouping_window_seconds INTEGER NOT NULL DEFAULT 30);
CREATE TABLE resources (
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
    component_id TEXT, confirmation_checks INTEGER NOT NULL DEFAULT 2, confirmation_interval INTEGER NOT NULL DEFAULT 30, expiry_alert_thresholds TEXT DEFAULT NULL, flap_detection_enabled INTEGER NOT NULL DEFAULT 1, flap_threshold INTEGER NOT NULL DEFAULT 4, flap_window_seconds INTEGER NOT NULL DEFAULT 600, flap_max_duration_minutes INTEGER NOT NULL DEFAULT 30, last_status_transition DATETIME, flap_started_at DATETIME, reminder_interval_minutes INTEGER NOT NULL DEFAULT 0, heartbeat_slug TEXT, heartbeat_interval INTEGER, heartbeat_grace INTEGER, last_ping_at DATETIME, keyword TEXT, keyword_mode TEXT, protocol_type TEXT, protocol_port INTEGER, host_id TEXT,
    FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE SET NULL
);
INSERT INTO resources VALUES('01M2FQRWWRHRF7CSZCNX0EFWDT','2026-09-14 10:35:28.536452 +0000 GMT m=+8.850998418','2026-09-14 10:35:28.55442 +0000 GMT m=+8.868966668','b4-mon','http',10,2,'http://127.0.0.1:1/','2026-09-14 10:35:50.718734 +0000 GMT m=+31.033188710','down',1,2,NULL,'',NULL,'',NULL,2,9,NULL,1,4,600,30,'2026-09-14 10:35:39.721192 +0000 GMT',NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01M2FQRWW6ZZJESQMFSR2086NQ');
INSERT INTO resources VALUES('01M2FQRWXHN8K1YBB4PFTKF0N6','2026-09-14 10:35:28.561779 +0000 GMT m=+8.876325376','2026-09-14 10:35:28.561779 +0000 GMT m=+8.876325376','b4-mon-up','http',60,5,'http://127.0.0.1:18094/api/v1/openapi.json',NULL,'pending',1,0,NULL,'',NULL,'',NULL,2,30,NULL,1,4,600,30,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
CREATE TABLE incidents (
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
INSERT INTO incidents VALUES('01M2FQSJHZSK4QN498H52HXG0G','2026-09-14 10:35:50.71928 +0000 GMT m=+31.033735210','2026-09-14 10:35:50.71928 +0000 GMT m=+31.033735210','01M2FQRWWRHRF7CSZCNX0EFWDT','HTTP Request Failed',NULL,'2026-09-14 10:35:50.719242 +0000 GMT m=+31.033697043',X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31');
CREATE TABLE incident_event_steps (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    incident_id TEXT NOT NULL,
    step TEXT NOT NULL,
    message TEXT,
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);
INSERT INTO incident_event_steps VALUES('01M2FQSJHZXZGV9TWCJFKPFBW1','2026-09-14 10:35:50.720029 +0000 GMT m=+31.034483710','2026-09-14 10:35:50.720029 +0000 GMT m=+31.034483710','01M2FQSJHZSK4QN498H52HXG0G','detected','Incident detected: HTTP request failed');
CREATE TABLE incident_diagnostics (
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
    body_encoded INTEGER NOT NULL DEFAULT 0, icmp_available INTEGER DEFAULT NULL, icmp_reachable INTEGER DEFAULT NULL, icmp_rtt_ms INTEGER DEFAULT NULL, root_cause_hint TEXT NOT NULL DEFAULT '', keyword TEXT, keyword_mode TEXT, keyword_found BOOLEAN,
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);
INSERT INTO incident_diagnostics VALUES('01M2FQSJHZH310MH007MRZ7V6W','2026-09-14 10:35:50.719617 +0000 GMT m=+31.034071293','2026-09-14 10:35:50.722383 +0000 GMT m=+31.036838085','01M2FQSJHZSK4QN498H52HXG0G','HEAD','http://127.0.0.1:1/','{}',2,-1,'{}','',0,'HTTP Request Failed','request error: Head "http://127.0.0.1:1/": safenet: failed to connect to any resolved IP for 127.0.0.1:1','',0,0,0,0,0,0,1,1,0,'service_down',NULL,NULL,NULL);
CREATE TABLE notification_events (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    incident_id TEXT NOT NULL,
    type TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending', claim_owner TEXT, claimed_at DATETIME, processed_at DATETIME, last_error TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);
CREATE TABLE notification_channels (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config BLOB NOT NULL,
    enabled_by_default INTEGER NOT NULL DEFAULT 0
, last_sent_at    DATETIME NULL, last_failure_at DATETIME NULL, failures_24h    INTEGER  NOT NULL DEFAULT 0);
CREATE TABLE maintenances (
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
CREATE TABLE monitoring_activities (
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
INSERT INTO monitoring_activities VALUES('01M2FQS7T8PXB1NZG7X0PS1A8C','2026-09-14 10:35:39.720277 +0000 GMT m=+20.034777543','2026-09-14 10:35:39.720277 +0000 GMT m=+20.034777543','01M2FQRWWRHRF7CSZCNX0EFWDT','Check http - Status: down',0,2,X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31',0);
INSERT INTO monitoring_activities VALUES('01M2FQSJHYR44XA8480BQSJ9S5','2026-09-14 10:35:50.718198 +0000 GMT m=+31.032653126','2026-09-14 10:35:50.718198 +0000 GMT m=+31.032653126','01M2FQRWWRHRF7CSZCNX0EFWDT','Check http - Status: down',0,1,X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31',0);
CREATE TABLE status_page_settings (
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
, custom_domain_status      TEXT NOT NULL DEFAULT 'pending', custom_domain_ssl_status  TEXT NOT NULL DEFAULT 'none', custom_domain_dns_records TEXT NOT NULL DEFAULT '[]', logo_url_light  TEXT NOT NULL DEFAULT '', logo_url_dark   TEXT NOT NULL DEFAULT '', favicon_url     TEXT NOT NULL DEFAULT '', primary_color   TEXT NOT NULL DEFAULT '#4f46e5', theme_overrides TEXT NOT NULL DEFAULT '{}', umami_website_id TEXT NOT NULL DEFAULT '', umami_script_url TEXT NOT NULL DEFAULT '');
CREATE TABLE users (
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
INSERT INTO users VALUES('01M2FQRMEV65A3BN8R9E7EV32J','admin@ogoune.test','Administrator','$2a$12$OervuBakJWaWYzYw4g/E0.xYSjqZjMPGIBmicOCcaYfHbFQ94XwOK',1,0,0,'',NULL,'2026-09-14 10:35:28','2026-09-14 10:35:19.89967 +0000 GMT m=+0.214252751','2026-09-14 10:35:19.89967 +0000 GMT m=+0.214252751');
CREATE TABLE resource_tags (
    resource_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (resource_id, tag_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
INSERT INTO resource_tags VALUES('01M2FQRWWRHRF7CSZCNX0EFWDT','01M2FQRWWR7T15NJMDTRV0XHG7');
CREATE TABLE resource_notification_channels (
    resource_id TEXT NOT NULL,
    notification_channel_id TEXT NOT NULL,
    PRIMARY KEY (resource_id, notification_channel_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);
CREATE TABLE component_notification_channels (
    component_id TEXT NOT NULL,
    notification_channel_id TEXT NOT NULL,
    PRIMARY KEY (component_id, notification_channel_id),
    FOREIGN KEY (component_id) REFERENCES components(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);
CREATE TABLE maintenance_resources (
    maintenance_id TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    PRIMARY KEY (maintenance_id, resource_id),
    FOREIGN KEY (maintenance_id) REFERENCES maintenances(id) ON DELETE CASCADE,
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);
CREATE TABLE expiry_notification_logs (
    id          TEXT     PRIMARY KEY,
    resource_id TEXT     NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    expiry_type TEXT     NOT NULL CHECK (expiry_type IN ('ssl', 'domain')),
    threshold   INTEGER  NOT NULL,
    sent_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (resource_id, expiry_type, threshold)
);
CREATE TABLE api_keys (
    id           TEXT     NOT NULL PRIMARY KEY,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id      TEXT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT     NOT NULL,
    key_hash     TEXT     NOT NULL UNIQUE,
    key_prefix   TEXT     NOT NULL,
    scope        TEXT     NOT NULL DEFAULT 'read',
    expires_at   DATETIME,
    last_used_at DATETIME,
    last_used_ip TEXT     NOT NULL DEFAULT '',
    is_active    INTEGER  NOT NULL DEFAULT 1
);
CREATE TABLE resource_credentials (
    id          TEXT PRIMARY KEY,
    resource_id TEXT NOT NULL UNIQUE,
    username    TEXT,
    password    BLOB NOT NULL,
    options     BLOB,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);
CREATE TABLE sessions (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL,
    browser        TEXT NOT NULL DEFAULT '',
    os             TEXT NOT NULL DEFAULT '',
    ip             TEXT NOT NULL DEFAULT '',
    location       TEXT,
    last_active_at DATETIME NOT NULL,
    created_at     DATETIME NOT NULL,
    revoked_at     DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
INSERT INTO sessions VALUES('01M2FQRWVTHXT9RQ78P02F9YGB','01M2FQRMEV65A3BN8R9E7EV32J','Unknown','Unknown','[::1]:51226',NULL,'2026-09-14 10:35:51.15817 +0000 UTC','2026-09-14 10:35:28.506539 +0000 UTC',NULL);
CREATE TABLE two_factor_reset_tokens (
    token_hash  TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    expires_at  DATETIME NOT NULL,
    used_at     DATETIME,
    created_at  DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE escalation_policies (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    scope_kind  TEXT NOT NULL,
    scope_value TEXT NOT NULL,
    is_active   INTEGER NOT NULL DEFAULT 1,
    priority    INTEGER NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);
CREATE TABLE escalation_steps (
    id            TEXT PRIMARY KEY,
    policy_id     TEXT NOT NULL,
    step_order    INTEGER NOT NULL,
    delay_minutes INTEGER NOT NULL,
    channel_ids   TEXT NOT NULL DEFAULT '[]',
    FOREIGN KEY (policy_id) REFERENCES escalation_policies(id) ON DELETE CASCADE
);
CREATE TABLE uptime_daily_agg (
    resource_id   TEXT     NOT NULL,
    day           TEXT     NOT NULL,
    samples       INTEGER  NOT NULL DEFAULT 0,
    up            INTEGER  NOT NULL DEFAULT 0,
    degraded      INTEGER  NOT NULL DEFAULT 0,
    down          INTEGER  NOT NULL DEFAULT 0,
    uptime_ratio  REAL     NOT NULL DEFAULT 1.0,
    computed_at   DATETIME NOT NULL,
    PRIMARY KEY (resource_id, day)
);
CREATE TABLE incident_updates (
    id          TEXT     NOT NULL PRIMARY KEY,
    incident_id TEXT     NOT NULL,
    status      TEXT     NOT NULL,
    message     TEXT     NOT NULL DEFAULT '',
    posted_by   TEXT     NOT NULL DEFAULT '',
    posted_at   DATETIME NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);
INSERT INTO incident_updates VALUES('01M2FQSJJ0AXV8FS3G4WKQJ66S','01M2FQSJHZSK4QN498H52HXG0G','investigating','We are currently investigating this issue.','','2026-09-14 10:35:50.720622 +0000 UTC','2026-09-14 10:35:50.720623 +0000 GMT m=+31.035077585','2026-09-14 10:35:50.720623 +0000 GMT m=+31.035077585');
CREATE TABLE notifications (
    id          TEXT PRIMARY KEY,
    user_id     TEXT,
    category    TEXT NOT NULL,
    severity    TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT,
    deep_link   TEXT,
    payload     TEXT,
    occurred_at DATETIME NOT NULL,
    read_at     DATETIME,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
INSERT INTO notifications VALUES('01M2FQSJHZE6B74YYKB3YKZZXV',NULL,'incident','error','b4-mon is down',NULL,'/incidents/01M2FQSJHZSK4QN498H52HXG0G','{"incident_id":"01M2FQSJHZSK4QN498H52HXG0G","resource_id":"01M2FQRWWRHRF7CSZCNX0EFWDT"}','2026-09-14 10:35:50.71957 +0000 GMT m=+31.034024418',NULL,'2026-09-14 10:35:50.719695 +0000 GMT m=+31.034149335','2026-09-14 10:35:50.719695 +0000 GMT m=+31.034149335');
CREATE TABLE dashboards (
    id                 TEXT PRIMARY KEY,
    owner_id           TEXT NOT NULL,
    name               TEXT NOT NULL,
    scope              TEXT NOT NULL,
    widgets            TEXT NOT NULL,
    default_time_range TEXT NOT NULL,
    refresh_interval   TEXT NOT NULL,
    visibility         TEXT NOT NULL,
    created_at         DATETIME NOT NULL,
    updated_at         DATETIME NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE report_history (
    id                 TEXT PRIMARY KEY,
    period             TEXT NOT NULL,
    sent_at            DATETIME NOT NULL,
    status             TEXT NOT NULL,
    uptime_pct         REAL NOT NULL,
    incident_count     INTEGER NOT NULL,
    downtime_seconds   INTEGER NOT NULL,
    recipient_email    TEXT NOT NULL,
    resource_breakdown TEXT NOT NULL,
    created_at         DATETIME NOT NULL
);
CREATE TABLE report_settings (
    id              TEXT PRIMARY KEY,
    enabled         INTEGER NOT NULL,
    recipient_email TEXT NOT NULL,
    schedule        TEXT NOT NULL,
    scope           TEXT NOT NULL,
    last_sent_at    DATETIME,
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL
);
CREATE TABLE announcements (
    id          TEXT PRIMARY KEY,
    severity    TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    dismissible INTEGER NOT NULL,
    active      INTEGER NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);
CREATE TABLE hosts (
    id            TEXT PRIMARY KEY,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    name          TEXT NOT NULL,
    os            TEXT,
    agent_version TEXT,
    last_seen_at  DATETIME,
    last_cpu_pct  REAL,
    last_mem_pct  REAL,
    last_disk_pct REAL,
    last_net_in   INTEGER,
    last_net_out  INTEGER,
    last_disks    TEXT
);
INSERT INTO hosts VALUES('01M2FQRWW6ZZJESQMFSR2086NQ','2026-09-14 10:35:28.518312 +0000 GMT m=+8.832858668','2026-09-14 10:35:28.518312 +0000 GMT m=+8.832858668','b4-host',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
CREATE TABLE host_credentials (
    id           TEXT PRIMARY KEY,
    created_at   DATETIME NOT NULL,
    updated_at   DATETIME NOT NULL,
    host_id      TEXT NOT NULL,
    hash         TEXT NOT NULL,
    prefix       TEXT NOT NULL,
    is_active    INTEGER NOT NULL,
    last_used_at DATETIME
);
INSERT INTO host_credentials VALUES('01M2FQRWW64KR791ZD18WG5YYM','2026-09-14 10:35:28.518586 +0000 GMT m=+8.833132585','2026-09-14 10:35:28.518586 +0000 GMT m=+8.833132585','01M2FQRWW6ZZJESQMFSR2086NQ','36c840742442ae1cb61441b4abaa34943a9a06f1ae088252257e7b59a056081d','ag_live_7baf',1,NULL);
CREATE TABLE host_metrics (
    id         TEXT PRIMARY KEY,
    host_id    TEXT NOT NULL,
    sampled_at DATETIME NOT NULL,
    cpu_pct    REAL NOT NULL,
    mem_pct    REAL NOT NULL,
    net_in     INTEGER NOT NULL,
    net_out    INTEGER NOT NULL,
    disks      TEXT
);
CREATE TABLE host_alert_state (
    host_id       TEXT PRIMARY KEY REFERENCES hosts(id) ON DELETE CASCADE,
    state         TEXT NOT NULL,
    offline_since DATETIME,
    alerted       INTEGER NOT NULL,
    updated_at    DATETIME NOT NULL
);
CREATE TABLE notification_escalation_state (
    id                    TEXT PRIMARY KEY,
    last_digest_at        DATETIME,
    watermark_occurred_at DATETIME,
    updated_at            DATETIME NOT NULL
);
CREATE UNIQUE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_tags_created_at ON tags(created_at);
CREATE UNIQUE INDEX idx_components_name ON components(name);
CREATE INDEX idx_components_created_at ON components(created_at);
CREATE INDEX idx_resources_created_at ON resources(created_at);
CREATE INDEX idx_resources_type ON resources(type);
CREATE INDEX idx_resources_status ON resources(status);
CREATE INDEX idx_resources_component_id ON resources(component_id);
CREATE INDEX idx_incidents_created_at ON incidents(created_at);
CREATE INDEX idx_incidents_resource_id ON incidents(resource_id);
CREATE INDEX idx_incidents_cause ON incidents(cause);
CREATE INDEX idx_incidents_resolved_at ON incidents(resolved_at);
CREATE INDEX idx_incidents_started_at ON incidents(started_at);
CREATE INDEX idx_incident_event_steps_created_at ON incident_event_steps(created_at);
CREATE INDEX idx_incident_event_steps_incident_id ON incident_event_steps(incident_id);
CREATE INDEX idx_incident_event_steps_step ON incident_event_steps(step);
CREATE INDEX idx_incident_diagnostics_created_at ON incident_diagnostics(created_at);
CREATE UNIQUE INDEX idx_incident_diagnostics_incident_id ON incident_diagnostics(incident_id);
CREATE INDEX idx_incident_diagnostics_http_status_code ON incident_diagnostics(http_status_code);
CREATE INDEX idx_incident_diagnostics_failure_type ON incident_diagnostics(failure_type);
CREATE INDEX idx_notification_events_created_at ON notification_events(created_at);
CREATE INDEX idx_notification_events_incident_id ON notification_events(incident_id);
CREATE INDEX idx_notification_events_type ON notification_events(type);
CREATE INDEX idx_notification_channels_created_at ON notification_channels(created_at);
CREATE INDEX idx_notification_channels_type ON notification_channels(type);
CREATE INDEX idx_maintenances_created_at ON maintenances(created_at);
CREATE INDEX idx_maintenances_strategy ON maintenances(strategy);
CREATE INDEX idx_maintenances_status ON maintenances(status);
CREATE INDEX idx_maintenances_start_at ON maintenances(start_at);
CREATE INDEX idx_maintenances_end_at ON maintenances(end_at);
CREATE INDEX idx_maintenances_cron_expr ON maintenances(cron_expr);
CREATE INDEX idx_maintenances_effective_from ON maintenances(effective_from);
CREATE INDEX idx_maintenances_effective_until ON maintenances(effective_until);
CREATE INDEX idx_maintenances_started_at ON maintenances(started_at);
CREATE INDEX idx_maintenances_ended_at ON maintenances(ended_at);
CREATE INDEX idx_monitoring_activities_created_at ON monitoring_activities(created_at);
CREATE INDEX idx_monitoring_activities_resource_id ON monitoring_activities(resource_id);
CREATE INDEX idx_status_page_settings_created_at ON status_page_settings(created_at);
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_resource_tags_tag_id ON resource_tags(tag_id);
CREATE INDEX idx_resource_notification_channels_channel_id ON resource_notification_channels(notification_channel_id);
CREATE INDEX idx_component_notification_channels_channel_id ON component_notification_channels(notification_channel_id);
CREATE INDEX idx_maintenance_resources_resource_id ON maintenance_resources(resource_id);
CREATE INDEX idx_expiry_notification_logs_resource_id ON expiry_notification_logs(resource_id);
CREATE INDEX idx_expiry_notification_logs_expiry_type ON expiry_notification_logs(expiry_type);
CREATE INDEX idx_expiry_notification_logs_sent_at ON expiry_notification_logs(sent_at);
CREATE INDEX idx_monitoring_activities_resource_created
    ON monitoring_activities(resource_id, created_at);
CREATE INDEX idx_resources_last_transition
    ON resources(last_status_transition);
CREATE INDEX idx_incidents_resource_resolved
  ON incidents (resource_id, resolved_at, started_at DESC);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_active ON api_keys(user_id, is_active);
CREATE INDEX idx_notification_events_status_created_at
    ON notification_events(status, created_at);
CREATE INDEX idx_notification_events_claim_owner
    ON notification_events(claim_owner);
CREATE UNIQUE INDEX idx_resources_heartbeat_slug_unique
    ON resources (heartbeat_slug)
    WHERE heartbeat_slug IS NOT NULL;
CREATE INDEX idx_resources_heartbeat_missed
    ON resources (type, status, last_ping_at);
CREATE INDEX idx_resource_credentials_resource_id ON resource_credentials(resource_id);
CREATE INDEX idx_sessions_user_active ON sessions(user_id, revoked_at);
CREATE INDEX idx_two_factor_reset_tokens_user ON two_factor_reset_tokens(user_id);
CREATE INDEX idx_two_factor_reset_tokens_expires ON two_factor_reset_tokens(expires_at);
CREATE INDEX idx_escalation_policies_active ON escalation_policies(is_active, priority);
CREATE UNIQUE INDEX uniq_escalation_priority_active
    ON escalation_policies(priority) WHERE is_active = 1;
CREATE UNIQUE INDEX uniq_escalation_steps_order
    ON escalation_steps(policy_id, step_order);
CREATE INDEX idx_uptime_daily_agg_day ON uptime_daily_agg(day);
CREATE INDEX idx_incident_updates_incident_posted
    ON incident_updates(incident_id, posted_at DESC);
CREATE INDEX idx_notifications_user_occurred ON notifications(user_id, occurred_at DESC);
CREATE INDEX idx_notifications_occurred ON notifications(occurred_at DESC);
CREATE INDEX idx_dashboards_owner ON dashboards(owner_id);
CREATE INDEX idx_dashboards_updated ON dashboards(updated_at DESC);
CREATE UNIQUE INDEX idx_report_history_period ON report_history(period);
CREATE INDEX idx_report_history_sent ON report_history(sent_at DESC);
CREATE INDEX idx_announcements_active ON announcements(active, created_at DESC);
CREATE INDEX idx_hosts_last_seen_at ON hosts(last_seen_at);
CREATE INDEX idx_host_credentials_hash ON host_credentials(hash);
CREATE INDEX idx_host_credentials_host_id ON host_credentials(host_id);
CREATE INDEX idx_host_metrics_host_sampled ON host_metrics(host_id, sampled_at);
CREATE INDEX idx_resources_host_id ON resources(host_id);
COMMIT;
