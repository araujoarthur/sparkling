-- name: CreateServiceToken :one
INSERT INTO auth.service_tokens (identity_id, token)
VALUES ($1, $2)
RETURNING *;

-- name: GetActiveServiceTokenByIdentity :one
SELECT * FROM auth.service_tokens
WHERE identity_id = $1
AND   revoked_at  IS NULL;

-- name: GetServiceTokenByToken :one
SELECT * FROM auth.service_tokens
WHERE token      = $1
AND   revoked_at IS NULL;

-- name: RevokeServiceToken :exec
UPDATE auth.service_tokens
SET revoked_at = now()
WHERE id = $1;

-- name: RevokeAllServiceTokensByIdentity :exec
UPDATE auth.service_tokens
SET revoked_at = now()
WHERE identity_id = $1
AND   revoked_at  IS NULL;

-- name: ListActiveServiceTokens :many
SELECT * FROM auth.service_tokens
WHERE revoked_at IS NULL
ORDER BY issued_at DESC;

-- name: GetServiceTokenByID :one
SELECT * FROM auth.service_tokens
WHERE id = $1;