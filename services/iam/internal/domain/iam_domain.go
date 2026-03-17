// iam_domain.go provides shared helpers used across all domain services.
// It is the domain layer's equivalent of iam_repository.go — a common
// foundation that every service file in this package depends on.
package domain

import (
	"context"
	"fmt"
	"regexp"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
	"github.com/google/uuid"
)

// roleNameRegexRule validates that a role name is lowercase with no spaces.
var roleNameRegexRule = regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)

// permissionNameRegexRule validates that a permission name follows the scope:resource:action format.
// Each segment must be lowercase. Actions may be hyphenated (e.g. assign-role).
var permissionNameRegexRule = regexp.MustCompile(`^[a-z]+:[a-z]+:[a-z]+(-[a-z]+)*$`)

// hasPermission checks whether a principal holds a specific permission.
func hasPermission(ctx context.Context, store *repository.Store, principalID uuid.UUID, permission string) (bool, error) {
	permissions, err := store.Principals.GetPermissions(ctx, principalID)
	if err != nil {
		return false, fmt.Errorf("hasPermission: %w", err)
	}

	for _, p := range permissions {
		if p.Name == permission {
			return true, nil
		}
	}

	return false, nil
}

// validateRoleName returns an error if the name does not meet requirements.
func validateRoleName(name string) error {
	if name == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	if !roleNameRegexRule.MatchString(name) {
		return fmt.Errorf("role name must be lowercase with no spaces; hyphens are allowed")
	}

	return nil
}

// validatePermissionName returns an error if the name does not meet requirements.
func validatePermissionName(name string) error {
	if name == "" {
		return fmt.Errorf("permission name cannot be empty")
	}

	if !permissionNameRegexRule.MatchString(name) {
		return fmt.Errorf("permission name must be a lowercase string separated by two colons, ex.: scope:resource:action")
	}

	return nil
}
