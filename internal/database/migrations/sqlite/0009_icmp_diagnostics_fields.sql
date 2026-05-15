-- 0009: ICMP diagnostic enrichment fields for incident_diagnostics
ALTER TABLE incident_diagnostics ADD COLUMN icmp_available INTEGER DEFAULT NULL;
ALTER TABLE incident_diagnostics ADD COLUMN icmp_reachable INTEGER DEFAULT NULL;
ALTER TABLE incident_diagnostics ADD COLUMN icmp_rtt_ms INTEGER DEFAULT NULL;
ALTER TABLE incident_diagnostics ADD COLUMN root_cause_hint TEXT NOT NULL DEFAULT '';
