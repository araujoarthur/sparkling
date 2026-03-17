-- name: AssignRoleToPrincipal :one
INSERT INTO iam.principal_roles (principal_id, role_id, granted_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: RemoveRoleFromPrincipal :exec
DELETE FROM iam.principal_roles
WHERE principal_id = $1
AND   role_id      = $2;

-- name: ListRolesByPrincipal :many
SELECT r.* FROM iam.roles r
INNER JOIN iam.principal_roles pr ON pr.role_id = r.id
WHERE pr.principal_id = $1
ORDER BY r.name ASC;

-- name: ListPrincipalsByRole :many
SELECT p.* FROM iam.principals p
INNER JOIN iam.principal_roles pr ON pr.principal_id = p.id
WHERE pr.role_id  = $1
AND   p.is_active = true
ORDER BY p.created_at ASC;

-- name: PrincipalHasRole :one
SELECT EXISTS (
    SELECT 1 FROM iam.principal_roles
    WHERE principal_id = $1
    AND   role_id      = $2
) AS has_role;

-- name: ListPrincipalRolesWithGranter :many
SELECT
    pr.principal_id,
    pr.role_id,
    pr.granted_by,
    pr.created_at,
    r.name        AS role_name,
    g.external_id AS granted_by_external_id
FROM iam.principal_roles pr
INNER JOIN iam.roles      r ON r.id = pr.role_id
INNER JOIN iam.principals g ON g.id = pr.granted_by
WHERE pr.principal_id = $1
ORDER BY pr.created_at ASC;