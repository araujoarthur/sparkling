-- name: GetPrincipalByID :one
SELECT * FROM iam.principals
WHERE id = $1;

-- name: GetPrincipalByExternalID :one
SELECT * FROM iam.principals
WHERE external_id    = $1
AND   principal_type = $2;

-- name: ListPrincipals :many
SELECT * FROM iam.principals
ORDER BY created_at ASC;

-- name: ListPrincipalsByType :many
SELECT * FROM iam.principals
WHERE principal_type = $1
ORDER BY created_at ASC;

-- name: CreatePrincipal :one
INSERT INTO iam.principals (external_id, principal_type)
VALUES ($1, $2)
RETURNING *;

-- name: ActivatePrincipal :one
UPDATE iam.principals
SET is_active = true
WHERE id = $1
RETURNING *;

-- name: DeactivatePrincipal :one
UPDATE iam.principals
SET is_active = false
WHERE id = $1
RETURNING *;

-- name: DeletePrincipal :exec
DELETE FROM iam.principals
WHERE id = $1;

-- name: GetPrincipalPermissions :many
SELECT DISTINCT p.* FROM iam.permissions p
INNER JOIN iam.role_permissions  rp ON rp.permission_id = p.id
INNER JOIN iam.principal_roles   pr ON pr.role_id       = rp.role_id
WHERE pr.principal_id = $1
AND   EXISTS (
    SELECT 1 FROM iam.principals
    WHERE id        = $1
    AND   is_active = true
)
ORDER BY p.name ASC;