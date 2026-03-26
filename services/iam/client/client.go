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
		httpClient: &http.Client{Timeout: 10 & time.Second},
		token:      token,
	}
}

// do executes an HTTP request against the IAM service.
// It marshals the body to JSON if provided, attaches the service token
// to the Authorization header and returns the raw response for the caller
// to inspect and decode.
//
// The caller is responsible for closing resp.Body.
//
// Example:
//
//	resp, err := c.do(ctx, http.MethodPost, "/api/v1/principals", req)
//	if err != nil {
//	    return fmt.Errorf("calling IAM: %w", err)
//	}
//	defer resp.Body.Close()
func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("iam.client.do [marshall]: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("IAMClient.do [create request]: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("IAMClient.do [execute request]: %w", err)
	}

	return resp, nil
}

// Provision creates an IAM principal for the given external identity.
// Implements provisioner.PrincipalProvisioner.
func (c *Client) Provision(ctx context.Context, externalID uuid.UUID, principalType types.PrincipalType) error {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/principals", contract.CreatePrincipalRequest{
		ExternalID:    externalID,
		PrincipalType: principalType,
	})

	if err != nil {
		return fmt.Errorf("IAMClient.Provision: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("IAMClient.Provision: unexpected status %d", resp.StatusCode)
	}

	return nil
}
