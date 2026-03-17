-- name: AssignPermissionToRole :exec
INSERT INTO iam.role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: RemovePermissionFromRole :exec
DELETE FROM iam.role_permissions
WHERE role_id      = $1
AND   permission_id = $2;

-- name: RoleHasPermission :one
SELECT EXISTS (
    SELECT 1 FROM iam.role_permissions
    WHERE role_id       = $1
    AND   permission_id = $2
) AS has_permission;

-- name: ListRolesByPermission :many
SELECT r.* FROM iam.roles r
INNER JOIN iam.role_permissions rp ON rp.role_id = r.id
WHERE rp.permission_id = $1
ORDER BY r.name ASC;