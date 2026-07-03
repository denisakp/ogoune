-- name: CreateAnnouncement :one
INSERT INTO announcements (id, severity, title, description, dismissible, active, created_at, updated_at)
VALUES (sqlc.arg('id'), sqlc.arg('severity'), sqlc.arg('title'), sqlc.arg('description'), sqlc.arg('dismissible'), sqlc.arg('active'), sqlc.arg('created_at'), sqlc.arg('updated_at'))
RETURNING *;

-- name: ListActiveAnnouncements :many
SELECT id, severity, title, description, dismissible, active, created_at, updated_at
FROM announcements
WHERE active = sqlc.arg('active')
ORDER BY created_at DESC;

-- name: DeleteAnnouncement :execrows
DELETE FROM announcements WHERE id = sqlc.arg('id');
