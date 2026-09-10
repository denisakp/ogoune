-- 0032: Database health frozen onto an incident (spec 088). Written once when the
-- incident is created and never updated, so it answers "how was this database
-- when the check broke" -- a different question from resource_health, which is
-- overwritten on every check and answers "how is it right now".
--
-- All four are nullable and independent: a non-database incident leaves them all
-- null and renders exactly as before.
--
-- Single file with no .down counterpart, following 0011_keyword_fields. This
-- migrator loads every .sql in the directory, including .down.sql, and executes
-- them in the forward path -- a DROP TABLE IF EXISTS is harmless when it runs
-- first, but a column drop is not. ALTER-based migrations therefore ship without
-- a down file here.
--
-- Distinct version from 0031 on purpose: the migrator keys applied migrations by
-- the numeric prefix alone, so two migrations sharing a number risk the second
-- being skipped permanently if a run is interrupted between them.
ALTER TABLE incident_diagnostics ADD COLUMN IF NOT EXISTS db_connections_active BIGINT;
ALTER TABLE incident_diagnostics ADD COLUMN IF NOT EXISTS db_connections_max BIGINT;
ALTER TABLE incident_diagnostics ADD COLUMN IF NOT EXISTS db_longest_query_seconds DOUBLE PRECISION;
ALTER TABLE incident_diagnostics ADD COLUMN IF NOT EXISTS db_replication_lag_seconds DOUBLE PRECISION;
