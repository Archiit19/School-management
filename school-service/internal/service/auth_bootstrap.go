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

type authBootstrapClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func newAuthBootstrapClient(baseURL, token string) *authBootstrapClient {
	return &authBootstrapClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *authBootstrapClient) BootstrapSchool(schoolID uuid.UUID) error {
	payload := map[string]string{"school_id": schoolID.String()}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/v1/internal/bootstrap-school", c.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("auth-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth-service bootstrap returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *authBootstrapClient) FetchRoleID(schoolID uuid.UUID, roleName string) (uuid.UUID, error) {
	url := fmt.Sprintf(
		"%s/api/v1/internal/roles/by-name?school_id=%s&name=%s",
		c.baseURL, schoolID.String(), roleName,
	)
	resp, err := c.client.Get(url)
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("auth-service returned status %d for role %s", resp.StatusCode, roleName)
	}
	var role struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(role.ID)
}

func (c *authBootstrapClient) AssignUserRole(userID, schoolID, roleID uuid.UUID) error {
	payload := map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
		"role_id":   roleID.String(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/internal/user-roles", c.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Internal-Token", c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("auth assign role returned status %d", resp.StatusCode)
	}
	return nil
}
