-- Resource: flap detection and reminder fields
ALTER TABLE resources ADD COLUMN flap_detection_enabled INTEGER NOT NULL DEFAULT 1;
ALTER TABLE resources ADD COLUMN flap_threshold INTEGER NOT NULL DEFAULT 4;
ALTER TABLE resources ADD COLUMN flap_window_seconds INTEGER NOT NULL DEFAULT 600;
ALTER TABLE resources ADD COLUMN flap_max_duration_minutes INTEGER NOT NULL DEFAULT 30;
ALTER TABLE resources ADD COLUMN last_status_transition DATETIME;
ALTER TABLE resources ADD COLUMN flap_started_at DATETIME;
ALTER TABLE resources ADD COLUMN reminder_interval_minutes INTEGER NOT NULL DEFAULT 0;

-- Component: grouping window
ALTER TABLE components ADD COLUMN grouping_window_seconds INTEGER NOT NULL DEFAULT 30;

CREATE INDEX IF NOT EXISTS idx_monitoring_activities_resource_created
    ON monitoring_activities(resource_id, created_at);

CREATE INDEX IF NOT EXISTS idx_resources_last_transition
    ON resources(last_status_transition);
