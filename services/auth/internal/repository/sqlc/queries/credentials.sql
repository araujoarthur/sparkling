-- name: CreateCredential :one
INSERT INTO auth.credentials (identity_id, type, identifier, secret_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCredentialByTypeAndIdentifier :one
SELECT * FROM auth.credentials
WHERE type       = $1
AND   identifier = $2;

-- name: GetCredentialsByIdentity :many
SELECT * FROM auth.credentials
WHERE identity_id = $1;

-- name: GetCredentialByIdentityAndType :one
SELECT * FROM auth.credentials
WHERE identity_id = $1
AND   type        = $2;

-- name: UpdateCredentialLastUsed :exec
UPDATE auth.credentials
SET last_used_at = now()
WHERE id = $1;

-- name: UpdateCredentialSecret :exec
UPDATE auth.credentials
SET secret_hash = $2
WHERE id = $1;

-- name: DeleteCredential :exec
DELETE FROM auth.credentials
WHERE id = $1;

-- name: DeleteCredentialsByIdentity :exec
DELETE FROM auth.credentials
WHERE identity_id = $1;