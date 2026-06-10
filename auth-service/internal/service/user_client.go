package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Archiit19/School-management/auth-service/internal/config"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/google/uuid"
)

type userClient struct {
	*httpclient.Client
}

func newUserClient(cfg *config.Config) *userClient {
	return &userClient{Client: httpclient.New(cfg.UserServiceURL, cfg.InternalServiceToken)}
}

func (c *userClient) enabled() bool {
	return c.BaseURL != ""
}

func (c *userClient) GetByEmail(email string) (*model.User, error) {
	if !c.enabled() {
		return nil, errors.New("user service is not configured")
	}
	path := fmt.Sprintf("/internal/users/by-email?email=%s", url.QueryEscape(email))
	req, err := http.NewRequest(http.MethodGet, c.URL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}
	if err := httpclient.CheckStatus(resp, http.StatusOK, "user-service get by email"); err != nil {
		return nil, err
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
	req, err := http.NewRequest(http.MethodGet, c.URL("/internal/users/"+id.String()), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}
	if err := httpclient.CheckStatus(resp, http.StatusOK, "user-service get by id"); err != nil {
		return nil, err
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
	var user model.User
	if err := c.DoJSONExpect(http.MethodPost, "/internal/users", payload, &user, http.StatusCreated); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *userClient) UpdateProfile(userID uuid.UUID, req model.UpdateProfileRequest) (*model.User, error) {
	if !c.enabled() {
		return nil, errors.New("user service is not configured")
	}
	var user model.User
	path := "/internal/users/" + userID.String()
	if err := c.DoJSONExpect(http.MethodPatch, path, req, &user, http.StatusOK); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *userClient) DeleteProfile(userID uuid.UUID) error {
	if !c.enabled() {
		return errors.New("user service is not configured")
	}
	req, err := http.NewRequest(http.MethodDelete, c.URL("/internal/users/"+userID.String()), nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "user-service delete")
}
