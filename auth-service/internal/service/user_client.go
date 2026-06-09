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

	"github.com/Archiit19/School-management/auth-service/internal/config"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/google/uuid"
)

type userClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func newUserClient(cfg *config.Config) *userClient {
	return &userClient{
		baseURL: strings.TrimRight(cfg.UserServiceURL, "/"),
		token:   cfg.InternalServiceToken,
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *userClient) enabled() bool {
	return c.baseURL != ""
}

func (c *userClient) do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("X-Internal-Token", c.token)
	}
	return c.client.Do(req)
}

func (c *userClient) GetByEmail(email string) (*model.User, error) {
	if !c.enabled() {
		return nil, errors.New("user service is not configured")
	}
	u := fmt.Sprintf("%s/internal/users/by-email?email=%s", c.baseURL, url.QueryEscape(email))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user-service returned status %d: %s", resp.StatusCode, string(b))
	}
	var user model.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *userClient) GetByID(id uuid.UUID) (*model.User, error) {
	if !c.enabled() {
		return nil, errors.New("user service is not configured")
	}
	u := fmt.Sprintf("%s/internal/users/%s", c.baseURL, id.String())
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
		return nil, fmt.Errorf("not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}
	var user model.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *userClient) CreateProfile(name, email string, studentID *uuid.UUID) (*model.User, error) {
	if !c.enabled() {
		return nil, errors.New("user service is not configured")
	}
	payload := map[string]interface{}{
		"name":  name,
		"email": email,
	}
	if studentID != nil {
		payload["student_id"] = studentID.String()
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/users", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user-service create returned status %d: %s", resp.StatusCode, string(b))
	}
	var user model.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *userClient) UpdateProfile(userID uuid.UUID, req model.UpdateProfileRequest) (*model.User, error) {
	if !c.enabled() {
		return nil, errors.New("user service is not configured")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/internal/users/%s", c.baseURL, userID.String())
	httpReq, err := http.NewRequest(http.MethodPatch, u, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user-service update returned status %d: %s", resp.StatusCode, string(b))
	}
	var user model.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *userClient) DeleteProfile(userID uuid.UUID) error {
	if !c.enabled() {
		return errors.New("user service is not configured")
	}
	u := fmt.Sprintf("%s/internal/users/%s", c.baseURL, userID.String())
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user-service delete returned status %d", resp.StatusCode)
	}
	return nil
}
