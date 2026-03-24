-- name: CreateRefreshToken :one
INSERT INTO auth.refresh_tokens (identity_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM auth.refresh_tokens
WHERE token_hash = $1
AND   revoked_at IS NULL
AND   expires_at > now();

-- name: GetActiveRefreshTokensByIdentity :many
SELECT * FROM auth.refresh_tokens
WHERE identity_id = $1
AND   revoked_at  IS NULL
AND   expires_at  > now()
ORDER BY created_at DESC;

-- name: RevokeRefreshToken :exec
UPDATE auth.refresh_tokens
SET revoked_at = now()
WHERE id = $1;

-- name: RevokeAllRefreshTokensByIdentity :exec
UPDATE auth.refresh_tokens
SET revoked_at = now()
WHERE identity_id = $1
AND   revoked_at  IS NULL;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM auth.refresh_tokens
WHERE expires_at < now();

-- name: DeleteAllRefreshTokensByIdentity :exec
DELETE FROM auth.refresh_tokens
WHERE identity_id = $1;

-- name: GetRefreshTokenByID :one
SELECT * FROM auth.refresh_tokens
WHERE id = $1;