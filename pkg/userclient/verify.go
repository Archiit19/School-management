package userclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/google/uuid"
)

// Client provides read-only calls to user-service for other microservices.
type Client struct {
	*httpclient.Client
}

// New creates a user-service internal client.
func New(baseURL, token string) *Client {
	return &Client{Client: httpclient.New(baseURL, token)}
}

// ParentHasChild returns true when the parent profile lists the child user ID.
func (c *Client) ParentHasChild(ctx context.Context, parentID, childID uuid.UUID) (bool, error) {
	if c.BaseURL == "" {
		return false, fmt.Errorf("user-service is not configured")
	}
	path := fmt.Sprintf("/internal/users/has-child/%s/%s", parentID.String(), childID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(path), nil)
	if err != nil {
		return false, err
	}
	resp, err := c.DoContext(ctx, req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, httpclient.CheckStatus(resp, http.StatusOK, "user-service has-child")
}
