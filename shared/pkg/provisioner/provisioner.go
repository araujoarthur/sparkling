// provisioner.go defines the PrincipalProvisioner interface used by the auth
// service to create IAM principals for newly registered identities.
// The default implementation calls IAM directly.
// A future implementation may enqueue a message for async processing.
package provisioner

import (
	"context"

	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
	"github.com/google/uuid"
)

// PrincipalProvisioner is responsible for creating an IAM principal
// for a newly registered identity.
// The default implementation calls IAM directly (eager provisioning).
// A future implementation may enqueue a message for async processing
// (eventual consistency) without any changes to the auth domain layer.
type PrincipalProvisioner interface {
	// Provision creates an IAM principal for the given external identity.
	// externalID is the identity ID issued by the auth service.
	// A failure here is non-fatal — the auth service logs the error and
	// continues. A reconciliation job handles unprovisioned identities.
	Provision(ctx context.Context, externalID uuid.UUID, principalType types.PrincipalType) error
}
