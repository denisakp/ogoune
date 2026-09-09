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

-- name: DeleteHostEventsOlderThan :execrows
-- Retention. Deletion only, never decimation: a kernel event is rare and discrete,
-- and thinning them would purge the records this feature exists to keep.
DELETE FROM host_events WHERE occurred_at < $1;

-- name: DeleteHostEventsByHost :exec
DELETE FROM host_events WHERE host_id = $1;
