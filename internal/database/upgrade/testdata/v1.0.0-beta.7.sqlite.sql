PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE schema_migrations (
    version TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME NOT NULL
);
INSERT INTO schema_migrations VALUES('0000','schema_migrations','2026-10-01 10:32:10.092788 +0000 UTC');
INSERT INTO schema_migrations VALUES('0001','initial','2026-10-01 10:32:10.09407 +0000 UTC');
INSERT INTO schema_migrations VALUES('0002','indexes','2026-10-01 10:32:10.095602 +0000 UTC');
INSERT INTO schema_migrations VALUES('0003','confirmation_fields','2026-10-01 10:32:10.096575 +0000 UTC');
INSERT INTO schema_migrations VALUES('0004','expiry_notification_logs','2026-10-01 10:32:10.097421 +0000 UTC');
INSERT INTO schema_migrations VALUES('0005','intelligent_alerting_fields','2026-10-01 10:32:10.099864 +0000 UTC');
INSERT INTO schema_migrations VALUES('0006','live_monitor_indexes','2026-10-01 10:32:10.100111 +0000 UTC');
INSERT INTO schema_migrations VALUES('0007','api_keys','2026-10-01 10:32:10.100417 +0000 UTC');
INSERT INTO schema_migrations VALUES('0008','pending_notification_retry','2026-10-01 10:32:10.102283 +0000 UTC');
INSERT INTO schema_migrations VALUES('0009','icmp_diagnostics_fields','2026-10-01 10:32:10.103765 +0000 UTC');
INSERT INTO schema_migrations VALUES('0010','heartbeat_fields','2026-10-01 10:32:10.105218 +0000 UTC');
INSERT INTO schema_migrations VALUES('0011','keyword_fields','2026-10-01 10:32:10.106824 +0000 UTC');
INSERT INTO schema_migrations VALUES('0012','protocol_fields','2026-10-01 10:32:10.107508 +0000 UTC');
INSERT INTO schema_migrations VALUES('0013','resource_credentials','2026-10-01 10:32:10.107775 +0000 UTC');
INSERT INTO schema_migrations VALUES('0014','sessions.up','2026-10-01 10:32:10.108053 +0000 UTC');
INSERT INTO schema_migrations VALUES('0015','two_factor.up','2026-10-01 10:32:10.108296 +0000 UTC');
INSERT INTO schema_migrations VALUES('0016','escalation_policies.up','2026-10-01 10:32:10.10863 +0000 UTC');
INSERT INTO schema_migrations VALUES('0017','custom_domains.up','2026-10-01 10:32:10.108969 +0000 UTC');
INSERT INTO schema_migrations VALUES('0018','statuspage_domain_state.up','2026-10-01 10:32:10.110259 +0000 UTC');
INSERT INTO schema_migrations VALUES('0019','statuspage_branding.up','2026-10-01 10:32:10.112121 +0000 UTC');
INSERT INTO schema_migrations VALUES('0020','uptime_daily_agg.up','2026-10-01 10:32:10.112371 +0000 UTC');
INSERT INTO schema_migrations VALUES('0021','incident_updates.up','2026-10-01 10:32:10.112627 +0000 UTC');
INSERT INTO schema_migrations VALUES('0022','notification_channel_dispatch_stats.up','2026-10-01 10:32:10.113873 +0000 UTC');
INSERT INTO schema_migrations VALUES('0023','replace_ga_with_umami.up','2026-10-01 10:32:10.114751 +0000 UTC');
INSERT INTO schema_migrations VALUES('0024','notifications.up','2026-10-01 10:32:10.115005 +0000 UTC');
INSERT INTO schema_migrations VALUES('0025','dashboards.up','2026-10-01 10:32:10.115269 +0000 UTC');
INSERT INTO schema_migrations VALUES('0026','report_history.up','2026-10-01 10:32:10.115543 +0000 UTC');
INSERT INTO schema_migrations VALUES('0027','announcements.up','2026-10-01 10:32:10.115745 +0000 UTC');
INSERT INTO schema_migrations VALUES('0028','hosts.up','2026-10-01 10:32:10.116572 +0000 UTC');
INSERT INTO schema_migrations VALUES('0029','host_alert_state.up','2026-10-01 10:32:10.116806 +0000 UTC');
INSERT INTO schema_migrations VALUES('0030','notification_escalation_state.up','2026-10-01 10:32:10.116972 +0000 UTC');
INSERT INTO schema_migrations VALUES('0031','resource_health.up','2026-10-01 10:32:10.117154 +0000 UTC');
INSERT INTO schema_migrations VALUES('0032','incident_diagnostics_db_health','2026-10-01 10:32:10.118914 +0000 UTC');
INSERT INTO schema_migrations VALUES('0033','host_events.up','2026-10-01 10:32:10.119144 +0000 UTC');
INSERT INTO schema_migrations VALUES('0034','report_settings.up','2026-10-01 10:32:10.11935 +0000 UTC');
INSERT INTO schema_migrations VALUES('0035','incident_host_link','2026-10-01 10:32:10.120333 +0000 UTC');
INSERT INTO schema_migrations VALUES('0036','capability_declaration','2026-10-01 10:32:10.122117 +0000 UTC');
CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    color TEXT,
    description TEXT
);
INSERT INTO tags VALUES('01M3VGB8RXJAZ40CY7QSBW16WZ','2026-10-01 10:32:16.925574 +0000 GMT m=+6.868691876','2026-10-01 10:32:16.925574 +0000 GMT m=+6.868691876','fx',NULL,NULL);
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
INSERT INTO resources VALUES('01M3VGB8RXKWXS1TGGBNWEQ417','2026-10-01 10:32:16.925918 +0000 GMT m=+6.869035501','2026-10-01 10:32:16.965676 +0000 GMT m=+6.908793417','fx-mon','http',10,2,'http://127.0.0.1:1/','2026-10-01 10:32:28.128553 +0000 GMT m=+18.071560292','down',1,1,NULL,'',NULL,'',NULL,1,1,NULL,1,4,600,30,'2026-10-01 10:32:28.128553 +0000 GMT m=+18.071560292',NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01M3VGB2TYTSG93G3MGVRDKGQG');
INSERT INTO resources VALUES('01M3VGB8TFRPWE86ZE57RZ0QJ3','2026-10-01 10:32:16.975433 +0000 GMT m=+6.918550042','2026-10-01 10:32:16.975433 +0000 GMT m=+6.918550042','fx-mon-nohost','http',10,2,'http://127.0.0.1:1/','2026-10-01 10:32:28.128649 +0000 GMT m=+18.071655667','down',1,1,NULL,'',NULL,'',NULL,1,1,NULL,1,4,600,30,'2026-10-01 10:32:28.128649 +0000 GMT m=+18.071655667',NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
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
INSERT INTO incidents VALUES('01M3VGBKQ1YMGWHJ4HVPD28YHK','2026-10-01 10:32:28.129798 +0000 GMT m=+18.072804959','2026-10-01 10:32:28.129798 +0000 GMT m=+18.072804959','01M3VGB8TFRPWE86ZE57RZ0QJ3','HTTP Request Failed',NULL,'2026-10-01 10:32:28.129762 +0000 GMT m=+18.072769459',X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31',NULL,1,NULL,'no_machine');
INSERT INTO incidents VALUES('01M3VGBKQ1BJHX3P78KZW1REP4','2026-10-01 10:32:28.129947 +0000 GMT m=+18.072954167','2026-10-01 10:32:28.129947 +0000 GMT m=+18.072954167','01M3VGB8RXKWXS1TGGBNWEQ417','HTTP Request Failed',NULL,'2026-10-01 10:32:28.129693 +0000 GMT m=+18.072699917',X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31','01M3VGB2TYTSG93G3MGVRDKGQG',1,'{"kmsg":{"available":false,"reason":"platform"},"cgroup_oom":{"available":false,"reason":"platform"},"segfault":{"available":false,"reason":"platform"}}','declared');
CREATE TABLE incident_event_steps (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    incident_id TEXT NOT NULL,
    step TEXT NOT NULL,
    message TEXT,
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);
INSERT INTO incident_event_steps VALUES('01M3VGBKQ43KMNNNAKECYXKXK8','2026-10-01 10:32:28.132814 +0000 GMT m=+18.075820834','2026-10-01 10:32:28.132814 +0000 GMT m=+18.075820834','01M3VGBKQ1BJHX3P78KZW1REP4','detected','Incident detected: HTTP request failed');
INSERT INTO incident_event_steps VALUES('01M3VGBKQ3JN0GGWQ1YATVN37P','2026-10-01 10:32:28.131368 +0000 GMT m=+18.074375292','2026-10-01 10:32:28.131368 +0000 GMT m=+18.074375292','01M3VGBKQ1YMGWHJ4HVPD28YHK','detected','Incident detected: HTTP request failed');
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
INSERT INTO incident_diagnostics VALUES('01M3VGBKQ21SGDYWMKC8DAJS9B','2026-10-01 10:32:28.130369 +0000 GMT m=+18.073376459','2026-10-01 10:32:28.137449 +0000 GMT m=+18.080456417','01M3VGBKQ1YMGWHJ4HVPD28YHK','HEAD','http://127.0.0.1:1/','{}',2,-1,'{}','',0,'HTTP Request Failed','request error: Head "http://127.0.0.1:1/": safenet: failed to connect to any resolved IP for 127.0.0.1:1','',0,0,0,0,0,0,1,1,0,'service_down',NULL,NULL,NULL,NULL,NULL,NULL,NULL);
INSERT INTO incident_diagnostics VALUES('01M3VGBKQ2SF8KYVVQ669CQHHX','2026-10-01 10:32:28.130741 +0000 GMT m=+18.073747626','2026-10-01 10:32:28.137273 +0000 GMT m=+18.080279792','01M3VGBKQ1BJHX3P78KZW1REP4','HEAD','http://127.0.0.1:1/','{}',2,-1,'{}','',0,'HTTP Request Failed','request error: Head "http://127.0.0.1:1/": safenet: failed to connect to any resolved IP for 127.0.0.1:1','',0,0,0,0,0,0,1,1,0,'service_down',NULL,NULL,NULL,NULL,NULL,NULL,NULL);
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
INSERT INTO monitoring_activities VALUES('01M3VGBKPYVCSN9QB5TYBQMPN7','2026-10-01 10:32:28.126867 +0000 GMT m=+18.069874501','2026-10-01 10:32:28.126867 +0000 GMT m=+18.069874501','01M3VGB8RXKWXS1TGGBNWEQ417','Check http - Status: down',0,1,X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31',0);
INSERT INTO monitoring_activities VALUES('01M3VGBKPZYW19FMZAVBH1Q8RB','2026-10-01 10:32:28.127331 +0000 GMT m=+18.070338501','2026-10-01 10:32:28.127331 +0000 GMT m=+18.070338501','01M3VGB8TFRPWE86ZE57RZ0QJ3','Check http - Status: down',0,1,X'72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31',0);
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
INSERT INTO users VALUES('01M3VGB2A4SSJZGYFJS9G2JBNY','fixture@ogoune.test','Administrator','$2a$12$Kec52lutfy4rMvLHxlUPV.yOTDi8jWInpos4gzRJ5aMvm1xA2oEz6',1,0,0,'',NULL,'2026-10-01 10:32:10','2026-10-01 10:32:10.308958 +0000 GMT m=+0.252141501','2026-10-01 10:32:10.308958 +0000 GMT m=+0.252141501');
CREATE TABLE resource_tags (
    resource_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (resource_id, tag_id),
    FOREIGN KEY (resource_id) REFERENCES resources(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
INSERT INTO resource_tags VALUES('01M3VGB8RXKWXS1TGGBNWEQ417','01M3VGB8RXJAZ40CY7QSBW16WZ');
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
INSERT INTO sessions VALUES('01M3VGB2TKQ4XSCJEZXGA03PN8','01M3VGB2A4SSJZGYFJS9G2JBNY','Unknown','Unknown','[::1]:53226',NULL,'2026-10-01 10:32:29.331578 +0000 UTC','2026-10-01 10:32:10.835483 +0000 UTC',NULL);
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
INSERT INTO incident_updates VALUES('01M3VGBKQ5YNGPWNEZKDEHHTER','01M3VGBKQ1BJHX3P78KZW1REP4','investigating','We are currently investigating this issue.','','2026-10-01 10:32:28.133681 +0000 UTC','2026-10-01 10:32:28.133682 +0000 GMT m=+18.076688917','2026-10-01 10:32:28.133682 +0000 GMT m=+18.076688917');
INSERT INTO incident_updates VALUES('01M3VGBKQ57BX6HBPQMN4J3VPY','01M3VGBKQ1YMGWHJ4HVPD28YHK','investigating','We are currently investigating this issue.','','2026-10-01 10:32:28.133961 +0000 UTC','2026-10-01 10:32:28.133962 +0000 GMT m=+18.076968667','2026-10-01 10:32:28.133962 +0000 GMT m=+18.076968667');
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
INSERT INTO notifications VALUES('01M3VGBKQ2MXGR7GXW31RT6CFP',NULL,'incident','error','fx-mon is down',NULL,'/incidents/01M3VGBKQ1BJHX3P78KZW1REP4','{"incident_id":"01M3VGBKQ1BJHX3P78KZW1REP4","resource_id":"01M3VGB8RXKWXS1TGGBNWEQ417"}','2026-10-01 10:32:28.130687 +0000 GMT m=+18.073693709',NULL,'2026-10-01 10:32:28.13078 +0000 GMT m=+18.073787501','2026-10-01 10:32:28.13078 +0000 GMT m=+18.073787501');
INSERT INTO notifications VALUES('01M3VGBKQ2P4ZJJ3VTDSBJQ1CQ',NULL,'incident','error','fx-mon-nohost is down',NULL,'/incidents/01M3VGBKQ1YMGWHJ4HVPD28YHK','{"incident_id":"01M3VGBKQ1YMGWHJ4HVPD28YHK","resource_id":"01M3VGB8TFRPWE86ZE57RZ0QJ3"}','2026-10-01 10:32:28.130294 +0000 GMT m=+18.073300834',NULL,'2026-10-01 10:32:28.130444 +0000 GMT m=+18.073450959','2026-10-01 10:32:28.130444 +0000 GMT m=+18.073450959');
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
INSERT INTO hosts VALUES('01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:10.846451 +0000 GMT m=+0.789628542','2026-10-01 10:32:27.486946 +0000 GMT m=+17.429959376','fx-host','darwin 26.6.2','0.1.0','2026-10-01 10:32:27.486686 +0000 UTC',39.29173693100570119,80.92975616455078125,97.3957857077542002,174382770025,54870099561,'[{"mount":"/","used_pct":92.85756758402793},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.85756758402793},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.85756758402793},{"mount":"/System/Volumes/Update","used_pct":92.85756758402793},{"mount":"/System/Volumes/VM","used_pct":92.85756758402793},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34812339104089}]','{"kmsg":{"available":false,"reason":"platform"},"cgroup_oom":{"available":false,"reason":"platform"},"segfault":{"available":false,"reason":"platform"}}','2026-10-01 10:32:27.486686 +0000 UTC');
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
INSERT INTO host_credentials VALUES('01M3VGB2TYX5CS5YADYJG5G9WK','2026-10-01 10:32:10.846654 +0000 GMT m=+0.789831417','2026-10-01 10:32:10.846654 +0000 GMT m=+0.789831417','01M3VGB2TYTSG93G3MGVRDKGQG','ef3fa5091c82000e4a7f77652527d53b1d74bbfc97f24a80496c24cc66d0f7b4','ag_live_964e',1,'2026-10-01 10:32:11.393562 +0000 UTC');
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
INSERT INTO host_metrics VALUES('01M3VGB3DEWTHKBFJZYS16VKR2','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:11.438994 +0000 UTC',17.08433121801648724,79.90818023681640625,174350302093,54868540615,'[{"mount":"/","used_pct":92.85969601194333},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.85969601194333},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.85969601194333},{"mount":"/System/Volumes/Update","used_pct":92.85969601194333},{"mount":"/System/Volumes/VM","used_pct":92.85969601194333},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34826117255939}]');
INSERT INTO host_metrics VALUES('01M3VGB5DPH4CKN4W57SPWD2A2','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:13.494571 +0000 UTC',26.17779598371862093,79.87804412841796875,174358586421,54868667347,'[{"mount":"/","used_pct":92.859773891361},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.859773891361},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.859773891361},{"mount":"/System/Volumes/Update","used_pct":92.859773891361},{"mount":"/System/Volumes/VM","used_pct":92.859773891361},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34826117255939}]');
INSERT INTO host_metrics VALUES('01M3VGB7C7JKCFWZH03B201MAQ','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:15.495632 +0000 UTC',24.2730720632081578,79.69385782877604641,174358630829,54868684501,'[{"mount":"/","used_pct":92.85981531658318},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.85981531658318},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.85981531658318},{"mount":"/System/Volumes/Update","used_pct":92.85981531658318},{"mount":"/System/Volumes/VM","used_pct":92.85981531658318},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34826117255939}]');
INSERT INTO host_metrics VALUES('01M3VGB9AK746A4DG83N5N5GPE','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:17.491713 +0000 UTC',28.15656565463549654,79.9762725830078125,174360372696,54868774971,'[{"mount":"/","used_pct":92.85933975503261},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.85933975503261},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.85933975503261},{"mount":"/System/Volumes/Update","used_pct":92.85933975503261},{"mount":"/System/Volumes/VM","used_pct":92.85933975503261},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34826117255939}]');
INSERT INTO host_metrics VALUES('01M3VGBB9F6W2SFN8KV5MKTMV0','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:19.503155 +0000 UTC',35.41230640217645487,79.61076100667317235,174369047503,54869277093,'[{"mount":"/","used_pct":92.8601467183606},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.8601467183606},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.8601467183606},{"mount":"/System/Volumes/Update","used_pct":92.8601467183606},{"mount":"/System/Volumes/VM","used_pct":92.8601467183606},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34826117255939}]');
INSERT INTO host_metrics VALUES('01M3VGBD7RGYE045Z9TNRC7K89','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:21.496003 +0000 UTC',35.69620253373474129,80.03660837809245265,174371985582,54869810272,'[{"mount":"/","used_pct":92.86040189772919},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.86040189772919},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.86040189772919},{"mount":"/System/Volumes/Update","used_pct":92.86040189772919},{"mount":"/System/Volumes/VM","used_pct":92.86040189772919},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34826117255939}]');
INSERT INTO host_metrics VALUES('01M3VGBF65CEQ7R3VTWNNAM6HC','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:23.493436 +0000 UTC',44.40251572333920649,80.97375233968098484,174373581849,54869894354,'[{"mount":"/","used_pct":92.86002161418962},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.86002161418962},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.86002161418962},{"mount":"/System/Volumes/Update","used_pct":92.86002161418962},{"mount":"/System/Volumes/VM","used_pct":92.86002161418962},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34826117255939}]');
INSERT INTO host_metrics VALUES('01M3VGBH4P4EFTGQJFR034H37J','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:25.494626 +0000 UTC',39.29471032837182065,81.38682047526042141,174382716731,54870041122,'[{"mount":"/","used_pct":92.85804977361406},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.85804977361406},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.85804977361406},{"mount":"/System/Volumes/Update","used_pct":92.85804977361406},{"mount":"/System/Volumes/VM","used_pct":92.85804977361406},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34826117255939}]');
INSERT INTO host_metrics VALUES('01M3VGBK2YTP5K3WTZ7Y8Y5JBW','01M3VGB2TYTSG93G3MGVRDKGQG','2026-10-01 10:32:27.486686 +0000 UTC',39.29173693100570119,80.92975616455078125,174382770025,54870099561,'[{"mount":"/","used_pct":92.85756758402793},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23A8464","used_pct":97.3957857077542},{"mount":"/Library/Developer/CoreSimulator/Volumes/iOS_23B86","used_pct":97.3812414220074},{"mount":"/System/Volumes/Data","used_pct":92.85756758402793},{"mount":"/System/Volumes/Hardware","used_pct":4.08046875},{"mount":"/System/Volumes/Preboot","used_pct":92.85756758402793},{"mount":"/System/Volumes/Update","used_pct":92.85756758402793},{"mount":"/System/Volumes/VM","used_pct":92.85756758402793},{"mount":"/System/Volumes/iSCPreboot","used_pct":4.08046875},{"mount":"/System/Volumes/xarts","used_pct":4.08046875},{"mount":"/Users/yaovi/OrbStack","used_pct":88.34812339104089}]');
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
