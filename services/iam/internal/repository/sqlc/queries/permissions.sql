-- name: GetPermissionByID :one
SELECT * FROM iam.permissions
WHERE id = $1;

-- name: GetPermissionByName :one
SELECT * FROM iam.permissions
WHERE name = $1;

-- name: ListPermissions :many
SELECT * FROM iam.permissions
ORDER BY name ASC;

-- name: ListPermissionsByRole :many
SELECT p.* FROM iam.permissions p
INNER JOIN iam.role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1
ORDER BY p.name ASC;

-- name: CreatePermission :one
INSERT INTO iam.permissions (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM iam.permissions
WHERE id = $1;