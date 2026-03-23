package keyprovider

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type EnvKeyProvider struct {
	privateKeyEnv string // env var name holding base64 PEM
	publicKeyEnv  string // env var name holding base64 PEM
}

func NewEnvKeyProvider(privateKeyEnv, publicKeyEnv string) *EnvKeyProvider {
	return &EnvKeyProvider{
		privateKeyEnv: privateKeyEnv,
		publicKeyEnv:  publicKeyEnv,
	}
}

// PrivateKey loads and parses an RSA private key from a base64-encoded PEM environment variable.
func (kp *EnvKeyProvider) PrivateKey() (*rsa.PrivateKey, error) {
	raw := os.Getenv(kp.privateKeyEnv)
	if raw == "" {
		return nil, fmt.Errorf("EnvKeyProvider.PrivateKey: %s is not set", kp.privateKeyEnv)
	}

	pemData, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("EnvKeyProvider.PrivateKey [decode base64]: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("EnvKeyProvider.PrivateKey [decode pem]: failed to decode PEM block from %s", kp.privateKeyEnv)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("EnvKeyProvider.PrivateKey [parse]: %w", err)
	}

	return key, nil
}

// PublicKey loads and parses an RSA public key from a base64-encoded PEM environment variable.
func (kp *EnvKeyProvider) PublicKey() (*rsa.PublicKey, error) {
	raw := os.Getenv(kp.publicKeyEnv)
	if raw == "" {
		return nil, fmt.Errorf("EnvKeyProvider.PublicKey: %s is not set", kp.publicKeyEnv)
	}

	pemData, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("EnvKeyProvider.PublicKey [decode base64]: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("EnvKeyProvider.PublicKey [decode pem]: failed to decode PEM block from %s", kp.publicKeyEnv)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("EnvKeyProvider.PublicKey [parse]: %w", err)
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("EnvKeyProvider.PublicKey: provided key is not a public key")
	}

	return pub, nil
}
