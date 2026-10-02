PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE schema_migrations (
    version TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME NOT NULL
);
INSERT INTO schema_migrations VALUES('0000','schema_migrations','2026-10-02 14:15:28.722607 +0000 UTC');
INSERT INTO schema_migrations VALUES('0001','initial','2026-10-02 14:15:28.723577 +0000 UTC');
INSERT INTO schema_migrations VALUES('0002','indexes','2026-10-02 14:15:28.725409 +0000 UTC');
INSERT INTO schema_migrations VALUES('0003','confirmation_fields','2026-10-02 14:15:28.726633 +0000 UTC');
INSERT INTO schema_migrations VALUES('0004','expiry_notification_logs','2026-10-02 14:15:28.727238 +0000 UTC');
INSERT INTO schema_migrations VALUES('0005','intelligent_alerting_fields','2026-10-02 14:15:28.729567 +0000 UTC');
INSERT INTO schema_migrations VALUES('0006','live_monitor_indexes','2026-10-02 14:15:28.729759 +0000 UTC');
INSERT INTO schema_migrations VALUES('0007','api_keys','2026-10-02 14:15:28.730039 +0000 UTC');
INSERT INTO schema_migrations VALUES('0008','pending_notification_retry','2026-10-02 14:15:28.731525 +0000 UTC');
INSERT INTO schema_migrations VALUES('0009','icmp_diagnostics_fields','2026-10-02 14:15:28.732728 +0000 UTC');
INSERT INTO schema_migrations VALUES('0010','heartbeat_fields','2026-10-02 14:15:28.733887 +0000 UTC');
INSERT INTO schema_migrations VALUES('0011','keyword_fields','2026-10-02 14:15:28.735215 +0000 UTC');
INSERT INTO schema_migrations VALUES('0012','protocol_fields','2026-10-02 14:15:28.735801 +0000 UTC');
INSERT INTO schema_migrations VALUES('0013','resource_credentials','2026-10-02 14:15:28.736065 +0000 UTC');
INSERT INTO schema_migrations VALUES('0014','sessions.up','2026-10-02 14:15:28.73634 +0000 UTC');
INSERT INTO schema_migrations VALUES('0015','two_factor.up','2026-10-02 14:15:28.736606 +0000 UTC');
INSERT INTO schema_migrations VALUES('0016','escalation_policies.up','2026-10-02 14:15:28.736931 +0000 UTC');
INSERT INTO schema_migrations VALUES('0017','custom_domains.up','2026-10-02 14:15:28.737335 +0000 UTC');
INSERT INTO schema_migrations VALUES('0018','statuspage_domain_state.up','2026-10-02 14:15:28.738469 +0000 UTC');
INSERT INTO schema_migrations VALUES('0019','statuspage_branding.up','2026-10-02 14:15:28.740034 +0000 UTC');
INSERT INTO schema_migrations VALUES('0020','uptime_daily_agg.up','2026-10-02 14:15:28.740291 +0000 UTC');
INSERT INTO schema_migrations VALUES('0021','incident_updates.up','2026-10-02 14:15:28.740469 +0000 UTC');
INSERT INTO schema_migrations VALUES('0022','notification_channel_dispatch_stats.up','2026-10-02 14:15:28.741499 +0000 UTC');
INSERT INTO schema_migrations VALUES('0023','replace_ga_with_umami.up','2026-10-02 14:15:28.742283 +0000 UTC');
INSERT INTO schema_migrations VALUES('0024','notifications.up','2026-10-02 14:15:28.742558 +0000 UTC');
INSERT INTO schema_migrations VALUES('0025','dashboards.up','2026-10-02 14:15:28.742764 +0000 UTC');
INSERT INTO schema_migrations VALUES('0026','report_history.up','2026-10-02 14:15:28.742972 +0000 UTC');
INSERT INTO schema_migrations VALUES('0027','announcements.up','2026-10-02 14:15:28.743137 +0000 UTC');
INSERT INTO schema_migrations VALUES('0028','hosts.up','2026-10-02 14:15:28.743878 +0000 UTC');
INSERT INTO schema_migrations VALUES('0029','host_alert_state.up','2026-10-02 14:15:28.744079 +0000 UTC');
INSERT INTO schema_migrations VALUES('0030','notification_escalation_state.up','2026-10-02 14:15:28.744223 +0000 UTC');
INSERT INTO schema_migrations VALUES('0031','resource_health.up','2026-10-02 14:15:28.744388 +0000 UTC');
INSERT INTO schema_migrations VALUES('0032','incident_diagnostics_db_health','2026-10-02 14:15:28.745958 +0000 UTC');
INSERT INTO schema_migrations VALUES('0033','host_events.up','2026-10-02 14:15:28.746196 +0000 UTC');
INSERT INTO schema_migrations VALUES('0034','report_settings.up','2026-10-02 14:15:28.746342 +0000 UTC');
INSERT INTO schema_migrations VALUES('0035','incident_host_link','2026-10-02 14:15:28.74715 +0000 UTC');
INSERT INTO schema_migrations VALUES('0036','capability_declaration','2026-10-02 14:15:28.748671 +0000 UTC');
CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    color TEXT,
    description TEXT
);
INSERT INTO tags VALUES('01M3YFGX14X5YFFG747VJ8XVG9','2026-10-02 14:15:36.229029 +0000 GMT m=+7.537566459','2026-10-02 14:15:36.229029 +0000 GMT m=+7.537566459','fx',NULL,NULL);
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
INSERT INTO resources VALUES('01M3YFGX15NKYFHC328RB4DTE7','2026-10-02 14:15:36.22942 +0000 GMT m=+7.537957001','2026-10-02 14:15:36.263729 +0000 GMT m=+7.572265834','fx-mon','http',10,2,'http://127.0.0.1:1/','2026-10-02 14:15:47.755421 +0000 GMT m=+19.063892418','down',1,1,NULL,'',NULL,'',NULL,1,1,NULL,1,4,600,30,'2026-10-02 14:15:47.755421 +0000 GMT m=+19.063892418',NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01M3YFGQ3EKV5BAZA81GK4K7P6');
INSERT INTO resources VALUES('01M3YFGX2PKPC1H8SV6MSNK364','2026-10-02 14:15:36.278696 +0000 GMT m=+7.587233084','2026-10-02 14:15:36.278696 +0000 GMT m=+7.587233084','fx-mon-nohost','http',10,2,'http://127.0.0.1:1/','2026-10-02 14:15:47.755368 +0000 GMT m=+19.063839251','down',1,1,NULL,'',NULL,'',NULL,1,1,NULL,1,4,600,30,'2026-10-02 14:15:47.755368 +0000 GMT m=+19.063839251',NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
CREATE TABLE incidents (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    resource_id TEXT NOT NULL,
    cause TEXT NOT NULL DEFAULT 'unknown_failure',
    resolved_at DATETIME,
    started_at DATETIME NOT NULL,
    details BLOB, host_id TEXT, host_link_recorded INTEGER NOT NULL DEFAULT 0, host_capabilities TEXT, host_capabilities_state TEXT,
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE
);
INSERT INTO incidents VALUES('01M3YFH89CFFF3AV1QDYF3RJB3','2026-10-02 14:15:47.756379 +0000 GMT m=+19.064850001','2026-10-02 14:15:47.756379 +0000 GMT m=+19.064850001','01M3YFGX2PKPC1H8SV6MSNK364','HTTP Request Failed',NULL,'2026-10-02 14:15:47.756348 +0000 GMT m=+19.064819293',X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31',NULL,1,NULL,'no_machine');
INSERT INTO incidents VALUES('01M3YFH89CAVWWNF13BFCN0FYY','2026-10-02 14:15:47.756887 +0000 GMT m=+19.065358626','2026-10-02 14:15:47.756887 +0000 GMT m=+19.065358626','01M3YFGX15NKYFHC328RB4DTE7','HTTP Request Failed',NULL,'2026-10-02 14:15:47.756446 +0000 GMT m=+19.064917709',X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31','01M3YFGQ3EKV5BAZA81GK4K7P6',1,'{"kmsg":{"available":false,"reason":"platform"},"cgroup_oom":{"available":false,"reason":"platform"},"segfault":{"available":false,"reason":"platform"}}','declared');
CREATE TABLE incident_event_steps (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    incident_id TEXT NOT NULL,
    step TEXT NOT NULL,
    message TEXT,
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);
INSERT INTO incident_event_steps VALUES('01M3YFH89DSKQET1SEGWBH7NRW','2026-10-02 14:15:47.757374 +0000 GMT m=+19.065845501','2026-10-02 14:15:47.757374 +0000 GMT m=+19.065845501','01M3YFH89CFFF3AV1QDYF3RJB3','detected','Incident detected: HTTP request failed');
INSERT INTO incident_event_steps VALUES('01M3YFH89EACD9BRHY0HQSNW94','2026-10-02 14:15:47.758965 +0000 GMT m=+19.067436043','2026-10-02 14:15:47.758965 +0000 GMT m=+19.067436043','01M3YFH89CAVWWNF13BFCN0FYY','detected','Incident detected: HTTP request failed');
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
    body_encoded INTEGER NOT NULL DEFAULT 0, icmp_available INTEGER DEFAULT NULL, icmp_reachable INTEGER DEFAULT NULL, icmp_rtt_ms INTEGER DEFAULT NULL, root_cause_hint TEXT NOT NULL DEFAULT '', keyword TEXT, keyword_mode TEXT, keyword_found BOOLEAN, db_connections_active INTEGER, db_connections_max INTEGER, db_longest_query_seconds REAL, db_replication_lag_seconds REAL,
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);
INSERT INTO incident_diagnostics VALUES('01M3YFH89C9KZAGXXMMF8223JS','2026-10-02 14:15:47.756812 +0000 GMT m=+19.065283668','2026-10-02 14:15:47.76135 +0000 GMT m=+19.069821126','01M3YFH89CFFF3AV1QDYF3RJB3','HEAD','http://127.0.0.1:1/','{}',2,-1,'{}','',0,'HTTP Request Failed','request error: Head "http://127.0.0.1:1/": safenet: failed to connect to any resolved IP for 127.0.0.1:1','',0,0,0,0,0,0,1,1,0,'service_down',NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO incident_diagnostics VALUES('01M3YFH89DQ8WBXBPPJFPE50YM','2026-10-02 14:15:47.757736 +0000 GMT m=+19.066206918','2026-10-02 14:15:47.761362 +0000 GMT m=+19.069832959','01M3YFH89CAVWWNF13BFCN0FYY','HEAD','http://127.0.0.1:1/','{}',2,-1,'{}','',0,'HTTP Request Failed','request error: Head "http://127.0.0.1:1/": safenet: failed to connect to any resolved IP for 127.0.0.1:1','',0,0,0,0,0,0,1,1,0,'service_down',NULL,NULL,NULL,NULL,NULL,NULL,NULL);
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
INSERT INTO notification_channels VALUES('01M3YFGX2FNWN75QH45QCG4PBZ','2026-10-02 14:15:36.27181 +0000 GMT m=+7.580346709','2026-10-02 14:15:36.27181 +0000 GMT m=+7.580346709','fx-mail','smtp',X'645a736d596271664e6f4a464d6975736b4d4349374a67774343475142636f4f623430614f4e2f555a446f336d4765473230546c77377536716b5862614842494847475050744f33586c4d7866466738506457327630724c387239707573417049637531364962764c5577682b59445a65476e45626f694e787651777a6f2b76596f4c71614d5a352f694c4b5173464c44654e39422b676b3156487747733066',0,NULL,NULL,0);
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
INSERT INTO monitoring_activities VALUES('01M3YFH899QA8JG6DA3CRWPAZ8','2026-10-02 14:15:47.753736 +0000 GMT m=+19.062207543','2026-10-02 14:15:47.753736 +0000 GMT m=+19.062207543','01M3YFGX2PKPC1H8SV6MSNK364','Check http - Status: down',0,1,X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31',0);
INSERT INTO monitoring_activities VALUES('01M3YFH899W92TQ8C63XN4AGBE','2026-10-02 14:15:47.75396 +0000 GMT m=+19.062431709','2026-10-02 14:15:47.75396 +0000 GMT m=+19.062431709','01M3YFGX15NKYFHC328RB4DTE7','Check http - Status: down',0,1,X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31',0);
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
INSERT INTO users VALUES('01M3YFGNX10K9RRKVZMNWQYXPM','fixture@ogoune.test','Administrator','$2a$12$vvZF1XKAlaYLGfFtS4bXZOH3u6w65x4ZmE2VbpCuvzyyY4Hraa9F.',1,0,0,'',NULL,'2026-10-02 14:15:30','2026-10-02 14:15:28.929691 +0000 GMT m=+0.238270126','2026-10-02 14:15:28.929691 +0000 GMT m=+0.238270126');
CREATE TABLE resource_tags (
    resource_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (resource_id, tag_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
INSERT INTO resource_tags VALUES('01M3YFGX15NKYFHC328RB4DTE7','01M3YFGX14X5YFFG747VJ8XVG9');
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
INSERT INTO sessions VALUES('01M3YFGQ33EM6M4CQDWN66WE7N','01M3YFGNX10K9RRKVZMNWQYXPM','Unknown','Unknown','[::1]:55878',NULL,'2026-10-02 14:15:48.603785 +0000 UTC','2026-10-02 14:15:30.147858 +0000 UTC',NULL);
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
INSERT INTO incident_updates VALUES('01M3YFH89EMJR0J5HVDNTM1XR1','01M3YFH89CFFF3AV1QDYF3RJB3','investigating','We are currently investigating this issue.','','2026-10-02 14:15:47.758485 +0000 UTC','2026-10-02 14:15:47.758486 +0000 GMT m=+19.066957751','2026-10-02 14:15:47.758486 +0000 GMT m=+19.066957751');
INSERT INTO incident_updates VALUES('01M3YFH89FMPNQQ1CWBS84DJZ5','01M3YFH89CAVWWNF13BFCN0FYY','investigating','We are currently investigating this issue.','','2026-10-02 14:15:47.759529 +0000 UTC','2026-10-02 14:15:47.759529 +0000 GMT m=+19.068000584','2026-10-02 14:15:47.759529 +0000 GMT m=+19.068000584');
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
INSERT INTO notifications VALUES('01M3YFH89CTGWWKBRV2ZME9XJV',NULL,'incident','error','fx-mon-nohost is down',NULL,'/incidents/01M3YFH89CFFF3AV1QDYF3RJB3','{"incident_id":"01M3YFH89CFFF3AV1QDYF3RJB3","resource_id":"01M3YFGX2PKPC1H8SV6MSNK364"}','2026-10-02 14:15:47.756745 +0000 GMT m=+19.065216626',NULL,'2026-10-02 14:15:47.75692 +0000 GMT m=+19.065391584','2026-10-02 14:15:47.75692 +0000 GMT m=+19.065391584');
INSERT INTO notifications VALUES('01M3YFH89D5Y11FE9TNXSECM9S',NULL,'incident','error','fx-mon is down',NULL,'/incidents/01M3YFH89CAVWWNF13BFCN0FYY','{"incident_id":"01M3YFH89CAVWWNF13BFCN0FYY","resource_id":"01M3YFGX15NKYFHC328RB4DTE7"}','2026-10-02 14:15:47.757697 +0000 GMT m=+19.066168418',NULL,'2026-10-02 14:15:47.757818 +0000 GMT m=+19.066289501','2026-10-02 14:15:47.757818 +0000 GMT m=+19.066289501');
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
, capabilities TEXT, capabilities_at DATETIME);
INSERT INTO hosts VALUES('01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:30.158313 +0000 GMT m=+1.466885084','2026-10-02 14:15:46.853864 +0000 GMT m=+18.162340126','fx-host','darwin 26.6.2','0.1.0','2026-10-02 14:15:46.853322 +0000 UTC',10.35058430557581666,69.25532023111979641,97.3957857077542002,181858530114,58563456340,'[{"mount":"/","used_pct":94.24878147501843},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24878147501843},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24878147501843},{"mount":"/System/Volumes/Update","used_pct":94.24878147501843},{"mount":"/System/Volumes/VM","used_pct":94.24878147501843},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]','{"kmsg":{"available":false,"reason":"platform"},"cgroup_oom":{"available":false,"reason":"platform"},"segfault":{"available":false,"reason":"platform"}}','2026-10-02 14:15:46.853322 +0000 UTC');
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
INSERT INTO host_credentials VALUES('01M3YFGQ3ESAR7Q4E3G7SXY81Q','2026-10-02 14:15:30.158475 +0000 GMT m=+1.467047126','2026-10-02 14:15:30.158475 +0000 GMT m=+1.467047126','01M3YFGQ3EKV5BAZA81GK4K7P6','ec8702e4754ca01e293a7eeba5f10085decec8a1655db955a3f334faa4df4f0a','ag_live_6a05',1,'2026-10-02 14:15:30.757384 +0000 UTC');
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
INSERT INTO host_metrics VALUES('01M3YFGQQBTKQYCJAKJ2T3WD78','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:30.795283 +0000 UTC',17.32110823440084601,69.73012288411457859,181858381539,58563227930,'[{"mount":"/","used_pct":94.24870359560074},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24870359560074},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24870359560074},{"mount":"/System/Volumes/Update","used_pct":94.24870359560074},{"mount":"/System/Volumes/VM","used_pct":94.24870359560074},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
INSERT INTO host_metrics VALUES('01M3YFGSPM8PYGYS15NPYDJVMX','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:32.820685 +0000 UTC',19.47890818867162465,69.843292236328125,181858397598,58563253084,'[{"mount":"/","used_pct":94.24873093624738},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24873093624738},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24873093624738},{"mount":"/System/Volumes/Update","used_pct":94.24873093624738},{"mount":"/System/Volumes/VM","used_pct":94.24873093624738},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
INSERT INTO host_metrics VALUES('01M3YFGVN4J72BWXJ9JWFHKY0E','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:34.820982 +0000 UTC',25.65816966043122349,69.8413848876953125,181858409785,58563263122,'[{"mount":"/","used_pct":94.24871519466296},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24871519466296},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24871519466296},{"mount":"/System/Volumes/Update","used_pct":94.24871519466296},{"mount":"/System/Volumes/VM","used_pct":94.24871519466296},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
INSERT INTO host_metrics VALUES('01M3YFGXMQ86BWP149DYNF3NX8','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:36.855861 +0000 UTC',12.70103093484773459,68.0477142333984375,181858442902,58563307170,'[{"mount":"/","used_pct":94.24878478903621},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24878478903621},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24878478903621},{"mount":"/System/Volumes/Update","used_pct":94.24878478903621},{"mount":"/System/Volumes/VM","used_pct":94.24878478903621},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
INSERT INTO host_metrics VALUES('01M3YFGZJXTQYSP9F2YSF7MMME','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:38.845715 +0000 UTC',11.46095717365117749,68.02399953206379734,181858466614,58563393876,'[{"mount":"/","used_pct":94.24875496287625},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24875496287625},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24875496287625},{"mount":"/System/Volumes/Update","used_pct":94.24875496287625},{"mount":"/System/Volumes/VM","used_pct":94.24875496287625},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
INSERT INTO host_metrics VALUES('01M3YFH1H8YWSXVGGAGBYRPXBG','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:40.8409 +0000 UTC',10.07133865029385334,68.13475290934245266,181858488250,58563421855,'[{"mount":"/","used_pct":94.24875910539846},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24875910539846},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24875910539846},{"mount":"/System/Volumes/Update","used_pct":94.24875910539846},{"mount":"/System/Volumes/VM","used_pct":94.24875910539846},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
INSERT INTO host_metrics VALUES('01M3YFH3G7C0WMEY596W7E6Q5Y','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:42.855837 +0000 UTC',11.19867275394336303,68.36414337158203125,181858500447,58563432018,'[{"mount":"/","used_pct":94.24879555959397},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24879555959397},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24879555959397},{"mount":"/System/Volumes/Update","used_pct":94.24879555959397},{"mount":"/System/Volumes/VM","used_pct":94.24879555959397},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
INSERT INTO host_metrics VALUES('01M3YFH5EKMN2FWDG2T8478309','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:44.851703 +0000 UTC',10.68286551726001576,70.03847757975260891,181858517382,58563445026,'[{"mount":"/","used_pct":94.24877816100067},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24877816100067},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24877816100067},{"mount":"/System/Volumes/Update","used_pct":94.24877816100067},{"mount":"/System/Volumes/VM","used_pct":94.24877816100067},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
INSERT INTO host_metrics VALUES('01M3YFH7D5QNTPFV5HK4C5BA78','01M3YFGQ3EKV5BAZA81GK4K7P6','2026-10-02 14:15:46.853322 +0000 UTC',10.35058430557581666,69.25532023111979641,181858530114,58563456340,'[{"mount":"/","used_pct":94.24878147501843},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":94.24878147501843},{"mount":"/System/Volumes/Hardware","used_pct":4.078125},{"mount":"/System/Volumes/Preboot","used_pct":94.24878147501843},{"mount":"/System/Volumes/Update","used_pct":94.24878147501843},{"mount":"/System/Volumes/VM","used_pct":94.24878147501843},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.078125},{"mount":"/System/Volumes/xarts","used_pct":4.078125},{"mount":"/Users/yaovi/OrbStack","used_pct":91.38132326750453}]');
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
CREATE TABLE resource_health (
    resource_id             TEXT PRIMARY KEY REFERENCES resources(id) ON DELETE CASCADE,
    collected_at            DATETIME NOT NULL,
    connections_active      INTEGER,
    connections_max         INTEGER,
    longest_query_seconds   REAL,
    replication_lag_seconds REAL,
    privilege_limited       INTEGER NOT NULL,
    unsupported_version     INTEGER NOT NULL
);
CREATE TABLE host_events (
    id          TEXT PRIMARY KEY,
    host_id     TEXT NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    occurred_at DATETIME NOT NULL,
    kind        TEXT NOT NULL,
    source      TEXT NOT NULL,
    occurrences INTEGER NOT NULL,
    detail      TEXT
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
CREATE INDEX idx_host_events_host_occurred ON host_events(host_id, occurred_at DESC);
COMMIT;
