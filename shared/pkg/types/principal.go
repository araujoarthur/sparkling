package types

// PrincipalType represents the kind of entity that can be assigned roles.
type PrincipalType string

const (
	PrincipalTypeUser    PrincipalType = "user"
	PrincipalTypeService PrincipalType = "service"
)
