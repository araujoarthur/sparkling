-- name: GetRoleByID :one
SELECT * FROM iam.roles
WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM iam.roles
WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM iam.roles
ORDER BY name ASC;

-- name: CreateRole :one
INSERT INTO iam.roles (name, description, is_system)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateRole :one
UPDATE iam.roles
SET name        = $2,
    description = $3
WHERE id = $1
AND   is_system = false
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM iam.roles
WHERE id = $1
AND   is_system = false;