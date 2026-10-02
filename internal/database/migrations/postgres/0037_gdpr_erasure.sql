-- 0037: Erase a person's data on request (spec 095).
--
-- Notification channels gain a disabled state. A channel an erasure leaves
-- with no recipient is disabled, not deleted: disabled_at says when,
-- disabled_reason why (closed code: erasure). NULL on every existing channel,
-- which stays enabled; nothing changes on upgrade. The four send finders skip
-- disabled channels, so no send path reaches one.
--
-- erasure_records keeps one row per erasure: when, which operator (user id,
-- no FK -- the operator may be erased later and the record stays), what kinds
-- of items changed (counts), and a keyed one-way fingerprint of the address.
-- Never the address itself.

ALTER TABLE notification_channels ADD COLUMN IF NOT EXISTS disabled_at TIMESTAMPTZ;
ALTER TABLE notification_channels ADD COLUMN IF NOT EXISTS disabled_reason TEXT;

CREATE TABLE IF NOT EXISTS erasure_records (
    id                  TEXT PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL,
    operator_id         TEXT NOT NULL,
    subject_kind        TEXT NOT NULL,
    subject_fingerprint TEXT NOT NULL,
    changes             JSONB NOT NULL,
    manual_review       INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_erasure_records_fingerprint ON erasure_records(subject_fingerprint);
