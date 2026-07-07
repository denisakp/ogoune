-- name: GetResourceCredentialByResourceID :one
SELECT * FROM resource_credentials WHERE resource_id = $1;

-- name: UpsertResourceCredential :exec
INSERT INTO resource_credentials (
    id, resource_id, username, password, options, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (resource_id) DO UPDATE
SET username = EXCLUDED.username,
    password = EXCLUDED.password,
    options = EXCLUDED.options,
    updated_at = EXCLUDED.updated_at;

-- name: DeleteResourceCredentialByResourceID :execrows
DELETE FROM resource_credentials WHERE resource_id = $1;

-- name: ResourceCredentialExists :one
SELECT COUNT(*) FROM resource_credentials WHERE resource_id = $1;
