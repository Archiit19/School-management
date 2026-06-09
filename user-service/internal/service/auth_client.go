package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/google/uuid"
)

type authClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func newAuthClient(cfg *config.Config) *authClient {
	return &authClient{
		baseURL: strings.TrimRight(cfg.AuthServiceURL, "/"),
		token:   cfg.InternalServiceToken,
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *authClient) do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("X-Internal-Token", c.token)
	}
	return c.client.Do(req)
}

func (c *authClient) SetCredential(userID uuid.UUID, password string) error {
	body, _ := json.Marshal(map[string]string{
		"user_id":  userID.String(),
		"password": password,
	})
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/credentials", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth set credential status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (c *authClient) DeleteUserAuth(userID uuid.UUID) error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+"/internal/credentials/"+userID.String(), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth delete credential status %d", resp.StatusCode)
	}
	return nil
}

func (c *authClient) AssignUserRole(userID, schoolID, roleID uuid.UUID) error {
	body, _ := json.Marshal(map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
		"role_id":   roleID.String(),
	})
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/user-roles", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth assign role status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (c *authClient) UpdateUserRole(userID, schoolID, roleID uuid.UUID) error {
	body, _ := json.Marshal(map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
		"role_id":   roleID.String(),
	})
	req, err := http.NewRequest(http.MethodPatch, c.baseURL+"/internal/user-roles", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth update role status %d", resp.StatusCode)
	}
	return nil
}

func (c *authClient) RemoveUserRole(userID, schoolID uuid.UUID) error {
	body, _ := json.Marshal(map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
	})
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+"/internal/user-roles", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth remove role status %d", resp.StatusCode)
	}
	return nil
}

type userRoleMember struct {
	UserID   uuid.UUID `json:"user_id"`
	SchoolID uuid.UUID `json:"school_id"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleName string    `json:"role_name"`
}

func (c *authClient) GetUserRole(userID, schoolID uuid.UUID) (*userRoleMember, error) {
	u := fmt.Sprintf("%s/internal/user-roles/%s?school_id=%s", c.baseURL, userID.String(), schoolID.String())
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("role not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth get role status %d", resp.StatusCode)
	}
	var m userRoleMember
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *authClient) ListUserRoles(userID uuid.UUID) ([]userRoleMember, error) {
	u := fmt.Sprintf("%s/internal/user-roles/%s", c.baseURL, userID.String())
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth list roles status %d", resp.StatusCode)
	}
	var rows []userRoleMember
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *authClient) GetRoleByID(roleID uuid.UUID) (name string, err error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/roles/%s", c.baseURL, roleID.String()))
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
	u := fmt.Sprintf("%s/api/v1/internal/roles/by-name?school_id=%s&name=student", c.baseURL, url.QueryEscape(schoolID.String()))
	resp, err := http.Get(u)
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
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/roles/%s/fields", c.baseURL, roleID.String()))
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
