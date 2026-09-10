-- 0032: Database health frozen onto an incident (spec 088). See postgres/0032.
-- One column per ALTER TABLE; SQLite has no ADD COLUMN IF NOT EXISTS.
-- Single file with no .down counterpart, following 0011_keyword_fields.
ALTER TABLE incident_diagnostics ADD COLUMN db_connections_active INTEGER;
ALTER TABLE incident_diagnostics ADD COLUMN db_connections_max INTEGER;
ALTER TABLE incident_diagnostics ADD COLUMN db_longest_query_seconds REAL;
ALTER TABLE incident_diagnostics ADD COLUMN db_replication_lag_seconds REAL;
