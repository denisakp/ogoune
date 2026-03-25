-- Composite index for recent activities query (resource detail live view)
CREATE INDEX IF NOT EXISTS idx_monitoring_activities_resource_created
  ON monitoring_activities (resource_id, created_at DESC);

-- Partial index for active incident per resource (most efficient form for PostgreSQL)
CREATE INDEX IF NOT EXISTS idx_incidents_resource_active
  ON incidents (resource_id, started_at DESC)
  WHERE resolved_at IS NULL;
