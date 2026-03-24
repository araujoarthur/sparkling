package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	UserTokenDuration = 15 * time.Minute
)

// Sentinel Errors
var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("invalid token")
)

type Claims struct {
	Subject       uuid.UUID
	PrincipalType types.PrincipalType
	IssuedAt      time.Time
	ExpiresAt     time.Time
}

type jwtClaims struct {
	jwt.RegisteredClaims
	PrincipalType types.PrincipalType `json:"principal_type"`
}

// Issue signs and returns a JWT token for the given claims using RS256.
// User tokens expire after 15 minutes.
// Service tokens are non-expiring — their validity is controlled by
// the auth service via revocation rather than token expiry.
// Should only be called by the auth service or inetbctl.
func Issue(claims Claims, privateKey *rsa.PrivateKey) (string, error) {
	now := time.Now()

	registered := jwt.RegisteredClaims{
		Subject:  claims.Subject.String(),
		IssuedAt: jwt.NewNumericDate(now),
	}

	// only user tokens expire — service tokens are controlled via revocation
	if claims.PrincipalType == types.PrincipalTypeUser {
		registered.ExpiresAt = jwt.NewNumericDate(now.Add(UserTokenDuration))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims{
		RegisteredClaims: registered,
		PrincipalType:    claims.PrincipalType,
	})

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("Issue [sign]: %w", err)
	}

	return signed, nil
}

// Parse validates and parses a JWT token string using the given RSA public key.
// Returns ErrExpiredToken if the token has expired.
// Returns ErrInvalidToken if the token is malformed or the signature is invalid.
func Parse(tokenString string, publicKey *rsa.PublicKey) (Claims, error) {
	// parse and verify the token signature using the public key.
	// the callback validates the signing method before verification
	// to prevent algorithm substitution attacks (e.g. switching to "none").
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwtClaims{},
		func(token *jwt.Token) (any, error) {
			// This is a security check against the algorithm confusion attack.
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		},
		jwt.WithIssuedAt(),
		jwt.WithLeeway(5*time.Second), // if two services have slightly different system clocks, a token that expired 2 seconds ago on one machine might still be valid on another. 5 seconds is a safe margin.
	)
	if err != nil {
		// check for expiry specifically so callers can handle it differently
		// (e.g. prompt a token refresh instead of returning 401)
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, ErrExpiredToken
		}
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	// assert the claims into our internal jwtClaims type.
	// token.Valid is a final safety check — false if anything failed validation.
	parsed, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	// parse the subject back into a UUID — it was stored as a string in the JWT.
	subject, err := uuid.Parse(parsed.Subject)
	if err != nil {
		return Claims{}, fmt.Errorf("Parse [subject]: %w", err)
	}

	// map the internal jwtClaims back into the domain Claims struct.
	return Claims{
		Subject:       subject,
		PrincipalType: parsed.PrincipalType,
		IssuedAt:      parsed.IssuedAt.Time,
		ExpiresAt:     parsed.ExpiresAt.Time,
	}, nil
}
