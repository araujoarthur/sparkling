-- name: CreateIdentity :one
INSERT INTO auth.identities (id)
VALUES (gen_random_uuid())
RETURNING *;

-- name: GetIdentityByID :one
SELECT * FROM auth.identities
WHERE id = $1;

-- name: DeleteIdentity :exec
DELETE FROM auth.identities
WHERE id = $1;