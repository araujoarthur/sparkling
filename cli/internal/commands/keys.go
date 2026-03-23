// keys.go provides commands for managing RSA key pairs used in token signing.
package commands

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func KeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage RSA key pairs for token signing",
	}

	cmd.AddCommand(generateKeysCmd())

	return cmd
}

func generateKeysCmd() *cobra.Command {
	var out string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate an RSA key pair for token signing",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(out, 0755); err != nil {
				return fmt.Errorf("creating output directory: %w", err)
			}

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				return fmt.Errorf("generating key pair: %w", err)
			}

			// write private key
			privateFile, err := os.Create(filepath.Join(out, "private.pem"))
			if err != nil {
				return fmt.Errorf("creating private key file: %w", err)
			}
			defer privateFile.Close()

			if err := pem.Encode(privateFile, &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
			}); err != nil {
				return fmt.Errorf("writing private key: %w", err)
			}

			// write public key
			pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
			if err != nil {
				return fmt.Errorf("marshaling public key: %w", err)
			}

			publicFile, err := os.Create(filepath.Join(out, "public.pem"))
			if err != nil {
				return fmt.Errorf("creating public key file: %w", err)
			}
			defer publicFile.Close()

			if err := pem.Encode(publicFile, &pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: pubBytes,
			}); err != nil {
				return fmt.Errorf("writing public key: %w", err)
			}

			fmt.Printf("Key pair generated at %s\n", out)
			return nil
		},
	}

	cmd.Flags().StringVar(&out, "out", "./keys", "Output directory for the generated key pair")
	return cmd
}
