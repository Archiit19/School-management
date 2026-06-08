package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type userBootstrapClient struct {
	baseURL string
	client  *http.Client
}

func newUserBootstrapClient(baseURL string) *userBootstrapClient {
	return &userBootstrapClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *userBootstrapClient) BootstrapSchool(schoolID uuid.UUID) error {
	payload := map[string]string{"school_id": schoolID.String()}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/internal/bootstrap-school", c.baseURL)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user-service bootstrap returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *userBootstrapClient) FetchRoleID(schoolID uuid.UUID, roleName string) (uuid.UUID, error) {
	url := fmt.Sprintf(
		"%s/api/v1/internal/roles/by-name?school_id=%s&name=%s",
		c.baseURL,
		schoolID.String(),
		roleName,
	)
	resp, err := c.client.Get(url)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("user-service returned status %d for role %s", resp.StatusCode, roleName)
	}
	var role struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(role.ID)
}
