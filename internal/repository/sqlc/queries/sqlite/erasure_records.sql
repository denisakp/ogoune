-- Spec 095: one row per erasure. Never the address: a keyed fingerprint.

-- name: CreateErasureRecord :exec
INSERT INTO erasure_records (
    id, created_at, operator_id, subject_kind, subject_fingerprint, changes, manual_review
)
VALUES (
    sqlc.arg('id'), sqlc.arg('created_at'), sqlc.arg('operator_id'), sqlc.arg('subject_kind'),
    sqlc.arg('subject_fingerprint'), sqlc.arg('changes'), sqlc.arg('manual_review')
);

-- name: ListErasureRecordsByFingerprint :many
SELECT * FROM erasure_records
WHERE subject_fingerprint = sqlc.arg('subject_fingerprint')
ORDER BY created_at DESC;
