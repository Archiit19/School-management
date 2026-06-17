package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Archiit19/School-management/pkg/httpx"
	"github.com/google/uuid"
)

type authBootstrapClient struct {
	httpx.Client
}

func newAuthBootstrapClient(baseURL, token string) *authBootstrapClient {
	return &authBootstrapClient{Client: httpx.New(baseURL, token)}
}

func (c *authBootstrapClient) BootstrapSchool(schoolID uuid.UUID) error {
	resp, err := c.DoJSON(http.MethodPost, "/api/v1/internal/bootstrap-school", map[string]string{
		"school_id": schoolID.String(),
	}, nil)
	if err != nil {
		return fmt.Errorf("auth-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if err := httpx.CheckStatus(resp, http.StatusOK, "auth-service bootstrap"); err != nil {
		return err
	}
	return nil
}

func (c *authBootstrapClient) FetchRoleID(schoolID uuid.UUID, roleName string) (uuid.UUID, error) {
	path := fmt.Sprintf("/api/v1/internal/roles/by-name?school_id=%s&name=%s", schoolID.String(), roleName)
	resp, err := c.Get(path)
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if err := httpx.CheckStatus(resp, http.StatusOK, "auth-service role lookup"); err != nil {
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
	resp, err := c.DoJSON(http.MethodPost, "/internal/user-roles", map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
		"role_id":   roleID.String(),
	}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpx.CheckStatus(resp, http.StatusCreated, "auth assign role")
}
