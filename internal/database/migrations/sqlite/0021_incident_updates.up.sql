CREATE TABLE IF NOT EXISTS incident_updates (
    id          TEXT     NOT NULL PRIMARY KEY,
    incident_id TEXT     NOT NULL,
    status      TEXT     NOT NULL,
    message     TEXT     NOT NULL DEFAULT '',
    posted_by   TEXT     NOT NULL DEFAULT '',
    posted_at   DATETIME NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_incident_updates_incident_posted
    ON incident_updates(incident_id, posted_at DESC);
