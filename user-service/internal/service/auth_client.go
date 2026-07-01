package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/google/uuid"
)

type authClient struct {
	*httpclient.Client
}

func (c *authClient) SetCredential(ctx context.Context, userID uuid.UUID, password string) error {
	resp, err := c.DoJSONContext(ctx, http.MethodPost, "/internal/credentials", map[string]string{
		"user_id":  userID.String(),
		"password": password,
	}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "auth set credential")
}

func (c *authClient) DeleteUserAuth(ctx context.Context, userID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.URL("/internal/credentials/"+userID.String()), nil)
	if err != nil {
		return err
	}
	resp, err := c.DoContext(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "auth delete credential")
}

func (c *authClient) AssignUserRole(ctx context.Context, userID, schoolID, roleID uuid.UUID) error {
	resp, err := c.DoJSONContext(ctx, http.MethodPost, "/internal/user-roles", map[string]string{
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

func (c *authClient) UpdateUserRole(ctx context.Context, userID, schoolID, roleID uuid.UUID) error {
	resp, err := c.DoJSONContext(ctx, http.MethodPatch, "/internal/user-roles", map[string]string{
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

func (c *authClient) RemoveUserRole(ctx context.Context, userID, schoolID uuid.UUID) error {
	resp, err := c.DoJSONContext(ctx, http.MethodDelete, "/internal/user-roles", map[string]string{
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

func (c *authClient) GetUserRole(ctx context.Context, userID, schoolID uuid.UUID) (*userRoleMember, error) {
	path := fmt.Sprintf("/internal/user-roles/%s?school_id=%s", userID.String(), schoolID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoContext(ctx, req)
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

func (c *authClient) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]userRoleMember, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL("/internal/user-roles/"+userID.String()), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoContext(ctx, req)
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

func (c *authClient) GetRoleByID(ctx context.Context, roleID uuid.UUID) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(fmt.Sprintf("/api/v1/roles/%s", roleID.String())), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.DoContext(ctx, req)
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

func (c *authClient) StudentRoleID(ctx context.Context, schoolID uuid.UUID) (uuid.UUID, error) {
	path := fmt.Sprintf("/api/v1/internal/roles/by-name?school_id=%s&name=student", url.QueryEscape(schoolID.String()))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(path), nil)
	if err != nil {
		return uuid.Nil, err
	}
	resp, err := c.DoContext(ctx, req)
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

func (c *authClient) GetRoleFields(ctx context.Context, roleID uuid.UUID) ([]fieldDefinition, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(fmt.Sprintf("/api/v1/roles/%s/fields", roleID.String())), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoContext(ctx, req)
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
