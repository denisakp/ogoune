ALTER TABLE resources ADD COLUMN keyword TEXT;
ALTER TABLE resources ADD COLUMN keyword_mode TEXT;

ALTER TABLE incident_diagnostics ADD COLUMN keyword TEXT;
ALTER TABLE incident_diagnostics ADD COLUMN keyword_mode TEXT;
ALTER TABLE incident_diagnostics ADD COLUMN keyword_found BOOLEAN;
