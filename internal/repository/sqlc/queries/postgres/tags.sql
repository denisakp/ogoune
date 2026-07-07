-- name: CreateTag :one
INSERT INTO tags (id, created_at, updated_at, name, color, description)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: FindTagByID :one
SELECT * FROM tags WHERE id = $1;

-- name: FindTagsByIDs :many
SELECT * FROM tags WHERE id = ANY($1::text[]);

-- name: FindTagByName :one
SELECT * FROM tags WHERE name = $1;

-- name: ListTags :many
SELECT * FROM tags
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateTag :execrows
UPDATE tags
SET name = $2, color = $3, description = $4, updated_at = $5
WHERE id = $1;

-- name: DeleteTag :execrows
DELETE FROM tags WHERE id = $1;
