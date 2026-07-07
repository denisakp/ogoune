-- name: CreateTag :one
INSERT INTO tags (id, created_at, updated_at, name, color, description)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: FindTagByID :one
SELECT * FROM tags WHERE id = ?;

-- name: FindTagsByIDs :many
SELECT * FROM tags WHERE id IN (sqlc.slice('ids'));

-- name: FindTagByName :one
SELECT * FROM tags WHERE name = ?;

-- name: ListTags :many
SELECT * FROM tags
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateTag :execrows
UPDATE tags
SET name = ?, color = ?, description = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteTag :execrows
DELETE FROM tags WHERE id = ?;
