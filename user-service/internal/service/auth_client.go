package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/google/uuid"
)

type authClient struct {
	*httpclient.Client
}

func newAuthClient(cfg *config.Config) *authClient {
	return &authClient{Client: httpclient.New(cfg.AuthServiceURL, cfg.InternalServiceToken)}
}

func (c *authClient) SetCredential(userID uuid.UUID, password string) error {
	resp, err := c.DoJSON(http.MethodPost, "/internal/credentials", map[string]string{
		"user_id":  userID.String(),
		"password": password,
	}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "auth set credential")
}

func (c *authClient) DeleteUserAuth(userID uuid.UUID) error {
	req, err := http.NewRequest(http.MethodDelete, c.URL("/internal/credentials/"+userID.String()), nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "auth delete credential")
}

func (c *authClient) AssignUserRole(userID, schoolID, roleID uuid.UUID) error {
	resp, err := c.DoJSON(http.MethodPost, "/internal/user-roles", map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
		"role_id":   roleID.String(),
	}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusCreated, "auth assign role")
}

func (c *authClient) UpdateUserRole(userID, schoolID, roleID uuid.UUID) error {
	resp, err := c.DoJSON(http.MethodPatch, "/internal/user-roles", map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
		"role_id":   roleID.String(),
	}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "auth update role")
}

func (c *authClient) RemoveUserRole(userID, schoolID uuid.UUID) error {
	resp, err := c.DoJSON(http.MethodDelete, "/internal/user-roles", map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
	}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "auth remove role")
}

type userRoleMember struct {
	UserID   uuid.UUID `json:"user_id"`
	SchoolID uuid.UUID `json:"school_id"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleName string    `json:"role_name"`
}

func (c *authClient) GetUserRole(userID, schoolID uuid.UUID) (*userRoleMember, error) {
	path := fmt.Sprintf("/internal/user-roles/%s?school_id=%s", userID.String(), schoolID.String())
	req, err := http.NewRequest(http.MethodGet, c.URL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("role not found")
	}
	if err := httpclient.CheckStatus(resp, http.StatusOK, "auth get role"); err != nil {
		return nil, err
	}
	var m userRoleMember
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *authClient) ListUserRoles(userID uuid.UUID) ([]userRoleMember, error) {
	req, err := http.NewRequest(http.MethodGet, c.URL("/internal/user-roles/"+userID.String()), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := httpclient.CheckStatus(resp, http.StatusOK, "auth list roles"); err != nil {
		return nil, err
	}
	var rows []userRoleMember
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *authClient) GetRoleByID(roleID uuid.UUID) (string, error) {
	resp, err := c.HTTP.Get(c.URL(fmt.Sprintf("/api/v1/roles/%s", roleID.String())))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("role not found")
	}
	var role struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return "", err
	}
	return role.Name, nil
}

func (c *authClient) StudentRoleID(schoolID uuid.UUID) (uuid.UUID, error) {
	path := fmt.Sprintf("/api/v1/internal/roles/by-name?school_id=%s&name=student", url.QueryEscape(schoolID.String()))
	resp, err := c.HTTP.Get(c.URL(path))
	if err != nil {
		return uuid.Nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, errors.New("student role not found")
	}
	var role struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(role.ID)
}

type fieldDefinition struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
}

func (c *authClient) GetRoleFields(roleID uuid.UUID) ([]fieldDefinition, error) {
	resp, err := c.HTTP.Get(c.URL(fmt.Sprintf("/api/v1/roles/%s/fields", roleID.String())))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	var fields []fieldDefinition
	if err := json.NewDecoder(resp.Body).Decode(&fields); err != nil {
		return nil, err
	}
	return fields, nil
}
