-- name: GetResourceCredentialByResourceID :one
SELECT * FROM resource_credentials WHERE resource_id = ?;

-- name: UpsertResourceCredential :exec
INSERT INTO resource_credentials (
    id, resource_id, username, password, options, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (resource_id) DO UPDATE
SET username = excluded.username,
    password = excluded.password,
    options = excluded.options,
    updated_at = excluded.updated_at;

-- name: DeleteResourceCredentialByResourceID :execrows
DELETE FROM resource_credentials WHERE resource_id = ?;

-- name: ResourceCredentialExists :one
SELECT COUNT(*) FROM resource_credentials WHERE resource_id = ?;
