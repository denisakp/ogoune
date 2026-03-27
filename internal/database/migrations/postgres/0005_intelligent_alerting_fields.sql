-- Resource: flap detection and reminder fields
ALTER TABLE resources
    ADD COLUMN IF NOT EXISTS flap_detection_enabled BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS flap_threshold INTEGER NOT NULL DEFAULT 4,
    ADD COLUMN IF NOT EXISTS flap_window_seconds INTEGER NOT NULL DEFAULT 600,
    ADD COLUMN IF NOT EXISTS flap_max_duration_minutes INTEGER NOT NULL DEFAULT 30,
    ADD COLUMN IF NOT EXISTS last_status_transition TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS flap_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS reminder_interval_minutes INTEGER NOT NULL DEFAULT 0;

-- Component: grouping window
ALTER TABLE components
    ADD COLUMN IF NOT EXISTS grouping_window_seconds INTEGER NOT NULL DEFAULT 30;

CREATE INDEX IF NOT EXISTS idx_monitoring_activities_resource_created
    ON monitoring_activities(resource_id, created_at);

CREATE INDEX IF NOT EXISTS idx_resources_last_transition
    ON resources(last_status_transition);
