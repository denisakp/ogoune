-- name: InsertHostEvent :exec
INSERT INTO host_events (id, host_id, occurred_at, kind, source, occurrences, detail)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListHostEventsByHost :many
-- Newest first: an operator opening a host page wants what just happened, and the
-- index on (host_id, occurred_at DESC) serves exactly this.
SELECT * FROM host_events
WHERE host_id = $1
ORDER BY occurred_at DESC
LIMIT $2;

-- name: ListHostEventsInWindow :many
-- Events inside one incident's correlation window (spec 091). Half-open
-- [from, to): an event exactly on the far edge belongs to one window and not two.
-- Newest first, and bounded -- a storm-prone host can hold many rows in a
-- six-minute window, and neither the sentence nor the served list grows with it.
-- Served by the existing index on (host_id, occurred_at DESC); no new index.
SELECT * FROM host_events
WHERE host_id = $1 AND occurred_at >= $2 AND occurred_at < $3
ORDER BY occurred_at DESC
LIMIT $4;

-- name: DeleteHostEventsOlderThan :execrows
-- Retention. Deletion only, never decimation: a kernel event is rare and discrete,
-- and thinning them would purge the records this feature exists to keep.
DELETE FROM host_events WHERE occurred_at < $1;

-- name: DeleteHostEventsByHost :exec
DELETE FROM host_events WHERE host_id = $1;
