package iamclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/araujoarthur/intranetbackend/services/iam/contract"
	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
	"github.com/google/uuid"
)

// Client is an HTTP client for the IAM service.
// It attaches a service token to every outgoing request.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// New constructs a Client with the given base URL and service token.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		token:      token,
	}
}

// do executes an HTTP request against the IAM service.
// It marshals the body to JSON if provided, attaches the service token
// to the Authorization header and returns the raw response for the caller
// to inspect and decode.
// actingPrincipal is the UUID of the user on whose behalf the request is made.
// Pass uuid.Nil if the client is acting on its own behalf.
//
// The caller is responsible for closing resp.Body.
// Example:
//
//	resp, err := c.do(ctx, http.MethodPost, "/api/v1/principals", req, actorID)
//	if err != nil {
//	    return fmt.Errorf("calling IAM: %w", err)
//	}
//	defer resp.Body.Close()
func (c *Client) do(ctx context.Context, method, path string, body any, actingPrincipal uuid.UUID) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	// only set X-Principal-ID if acting on behalf of a specific principal
	if actingPrincipal != uuid.Nil {
		req.Header.Set("X-Principal-ID", actingPrincipal.String())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

// Provision creates an IAM principal for the given external identity.
// Implements provisioner.PrincipalProvisioner.
func (c *Client) Provision(ctx context.Context, externalID uuid.UUID, principalType types.PrincipalType) error {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/principals", contract.CreatePrincipalRequest{
		ExternalID:    externalID,
		PrincipalType: principalType,
	}, uuid.Nil)

	if err != nil {
		return fmt.Errorf("IAMClient.Provision: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("IAMClient.Provision: unexpected status %d", resp.StatusCode)
	}

	return nil
}

// HasPermission checks whether a principal holds a specific permission in IAM.
// Calls GET /api/v1/principals/{id}/permissions and checks if the permission
// is present in the response.
func (c *Client) HasPermission(ctx context.Context, principalID uuid.UUID, permission string) (bool, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/principals/%s/permissions", principalID), nil, principalID)
	if err != nil {
		return false, fmt.Errorf("IAMClient.HasPermission: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("IAMClient.HasPermission: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Data []contract.PermissionResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("IAMClient.HasPermission [decode]: %w", err)
	}

	for _, p := range result.Data {
		if p.Name == permission {
			return true, nil
		}
	}

	return false, nil
}
