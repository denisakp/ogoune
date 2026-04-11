-- 0012: Protocol monitor schema fields
ALTER TABLE resources ADD COLUMN protocol_type TEXT;
ALTER TABLE resources ADD COLUMN protocol_port INTEGER;
