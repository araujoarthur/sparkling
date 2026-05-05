package token

import (
	"crypto/rsa"

	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
	"github.com/google/uuid"
)

// IssueUserToken issues a short-lived access token for a human user.
// The token expires after UserTokenDuration (15 minutes).
func IssueUserToken(identityID uuid.UUID, privateKey *rsa.PrivateKey) (string, error) {
	return Issue(Claims{
		Subject:       identityID,
		PrincipalType: types.PrincipalTypeUser,
	}, privateKey)
}

// IssueServiceToken issues a non-expiring token for a service account.
// Validity is controlled by revocation in the auth service, not by expiry.
func IssueServiceToken(principalID uuid.UUID, privateKey *rsa.PrivateKey) (string, error) {
	return Issue(Claims{
		Subject:       principalID,
		PrincipalType: types.PrincipalTypeService,
	}, privateKey)
}
