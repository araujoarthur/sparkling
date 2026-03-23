// builtins.go defines constant permission names used across domain services.
// These must match the permission names seeded into the database exactly.
package domain

import (
	"fmt"
	"strings"
)

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

	// Principals permissions
	permissionIAMPrincipalsWrite  = "iam:principals:write"
	permissionIAMPrincipalsDelete = "iam:principals:delete"
)

// GetGrantRolePermission returns the grant permission name for the given role.
// Used to check whether a caller is allowed to assign or revoke a specific role.
// Format: iam:role-{rolename}:grant
func GetGrantRolePermission(roleName string) string {
	return fmt.Sprintf("iam:role-%s:grant", strings.ToLower(roleName))
}
