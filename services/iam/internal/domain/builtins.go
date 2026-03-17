// builtins.go defines constant permission names used across domain services.
// These must match the permission names seeded into the database exactly.
package domain

const (
	// Role permissions
	permissionIAMRolesWrite  = "iam:roles:write"
	permissionIAMRolesDelete = "iam:roles:delete"

	// Permission permissions
	permissionIAMPermissionsWrite  = "iam:permissions:write"
	permissionIAMPermissionsDelete = "iam:permissions:delete"

	// Role permission assignments
	permissionIAMRolePermissionsAssign = "iam:permissions:assign"
	permissionIAMRolePermissionsRevoke = "iam:permissions:revoke"
)
