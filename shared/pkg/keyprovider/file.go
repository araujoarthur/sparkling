package keyprovider

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type FileKeyProvider struct {
	privateKeyPath string
	publicKeyPath  string
}

func NewFileKeyProvider(privateKeyPath, publicKeyPath string) *FileKeyProvider {
	return &FileKeyProvider{
		privateKeyPath: privateKeyPath,
		publicKeyPath:  publicKeyPath,
	}
}

// PrivateKey loads and parses an RSA private key from a PEM file.
func (kp *FileKeyProvider) PrivateKey() (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(kp.privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("FileKeyProvider.PrivateKey [read]: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("FileKeyProvider.PrivateKey [decode]: failed to decode PEM block from %s", kp.privateKeyPath)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("FileKeyProvider.PrivateKey [parse]: %w", err)
	}

	return key, nil
}

// PublicKey loads and parses an RSA public key from a PEM file.
func (kp *FileKeyProvider) PublicKey() (*rsa.PublicKey, error) {
	data, err := os.ReadFile(kp.publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("FileKeyProvider.PublicKey [read]: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("FileKeyProvider.PublicKey [decode]: failed to decode PEM block from %s", kp.publicKeyPath)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("FileKeyProvider.PublicKey [parse]: %w", err)
	}

	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("FileKeyProvider.PublicKey: provided key is not a public key")
	}

	return pub, nil
}
