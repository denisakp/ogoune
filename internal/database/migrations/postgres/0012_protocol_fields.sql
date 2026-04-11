-- 0012: Protocol monitor schema fields
ALTER TABLE resources ADD COLUMN IF NOT EXISTS protocol_type TEXT;
ALTER TABLE resources ADD COLUMN IF NOT EXISTS protocol_port INTEGER;
