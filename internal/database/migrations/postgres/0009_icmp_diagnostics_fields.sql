-- 0009: ICMP diagnostic enrichment fields for incident_diagnostics
ALTER TABLE incident_diagnostics
    ADD COLUMN IF NOT EXISTS icmp_available BOOLEAN DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS icmp_reachable BOOLEAN DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS icmp_rtt_ms INTEGER DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS root_cause_hint TEXT NOT NULL DEFAULT '';
