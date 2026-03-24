// main.go is the composition root for the IAM service.
// It wires together configuration, database, domain, and HTTP layers.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/domain"
	"github.com/araujoarthur/intranetbackend/services/iam/internal/handler/rest"
	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
	"github.com/araujoarthur/intranetbackend/shared/pkg/database"
	"github.com/araujoarthur/intranetbackend/shared/pkg/keyprovider"
	"github.com/joho/godotenv"
)

func main() {
	// load environment variables from .env if present
	_ = godotenv.Load()

	// load public key for token validation
	keyProvider := keyprovider.NewFileKeyProvider(
		os.Getenv("PRIVATE_KEY_PATH"), // not used by IAM but FileKeyProvider requires it
		os.Getenv("PUBLIC_KEY_PATH"),
	)

	publicKey, err := keyProvider.PublicKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load public key: %v\n", err)
		os.Exit(1)
	}

	// connect to database
	pool, err := database.NewPool(context.Background(), database.Config{
		DSN: os.Getenv("IAM_DSN"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// build repository store
	store := repository.NewStore(pool)

	// build domain services
	roles := domain.NewRoleService(store)
	permissions := domain.NewPermissionService(store)
	rolePermissions := domain.NewRolePermissionService(store)
	principals := domain.NewPrincipalService(store)
	principalRoles := domain.NewPrincipalRoleService(store)

	// build HTTP server
	server := rest.NewServer(
		publicKey,
		roles,
		permissions,
		rolePermissions,
		principals,
		principalRoles,
	)

	// start listening
	addr := os.Getenv("IAM_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	fmt.Printf("IAM service listening on %s\n", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
