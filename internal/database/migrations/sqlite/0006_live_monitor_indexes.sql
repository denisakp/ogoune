-- Composite index for recent activities query (resource detail live view)
CREATE INDEX IF NOT EXISTS idx_monitoring_activities_resource_created
  ON monitoring_activities (resource_id, created_at DESC);

-- Full composite index (SQLite does not support partial indexes with WHERE clause)
CREATE INDEX IF NOT EXISTS idx_incidents_resource_resolved
  ON incidents (resource_id, resolved_at, started_at DESC);
