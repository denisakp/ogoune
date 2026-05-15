ALTER TABLE resources ADD COLUMN IF NOT EXISTS keyword TEXT;
ALTER TABLE resources ADD COLUMN IF NOT EXISTS keyword_mode TEXT;

ALTER TABLE incident_diagnostics ADD COLUMN IF NOT EXISTS keyword TEXT;
ALTER TABLE incident_diagnostics ADD COLUMN IF NOT EXISTS keyword_mode TEXT;
ALTER TABLE incident_diagnostics ADD COLUMN IF NOT EXISTS keyword_found BOOLEAN;
