package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Archiit19/School-management/auth-service/internal/config"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/google/uuid"
)

type schoolMembership struct {
	UserID   uuid.UUID `json:"user_id"`
	SchoolID uuid.UUID `json:"school_id"`
}

type schoolClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func newSchoolClient(cfg *config.Config) *schoolClient {
	return &schoolClient{
		baseURL: strings.TrimRight(cfg.SchoolServiceURL, "/"),
		token:   cfg.InternalServiceToken,
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (c *schoolClient) enabled() bool {
	return c.baseURL != "" && strings.TrimSpace(c.token) != ""
}

func (c *schoolClient) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Internal-Token", c.token)
	return c.client.Do(req)
}

func (c *schoolClient) CreateSchoolForUser(userID uuid.UUID, name, address, phone, email string) (*model.School, error) {
	if !c.enabled() {
		return nil, errors.New("school service is not configured")
	}

	payload := map[string]string{
		"user_id": userID.String(),
		"name":    name,
		"address": address,
		"phone":   phone,
		"email":   email,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/schools/with-admin", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("school-service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return nil, errors.New("school with this email already exists")
	}
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("school-service create returned status %d: %s", resp.StatusCode, string(b))
	}

	var school model.School
	if err := json.NewDecoder(resp.Body).Decode(&school); err != nil {
		return nil, err
	}
	return &school, nil
}

func (c *schoolClient) GetSchoolByEmail(email string) (*model.School, error) {
	if !c.enabled() {
		return nil, errors.New("school service is not configured")
	}

	url := fmt.Sprintf("%s/internal/schools/by-email?email=%s", c.baseURL, email)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("school-service returned status %d", resp.StatusCode)
	}

	var school model.School
	if err := json.NewDecoder(resp.Body).Decode(&school); err != nil {
		return nil, err
	}
	return &school, nil
}

func (c *schoolClient) GetSchoolByID(id uuid.UUID) (*model.School, error) {
	if !c.enabled() {
		return nil, errors.New("school service is not configured")
	}

	url := fmt.Sprintf("%s/internal/schools/%s", c.baseURL, id.String())
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("school-service returned status %d: %s", resp.StatusCode, string(body))
	}

	var school model.School
	if err := json.NewDecoder(resp.Body).Decode(&school); err != nil {
		return nil, err
	}
	return &school, nil
}

func (c *schoolClient) ListSchoolsForUser(userID uuid.UUID) ([]model.School, error) {
	if !c.enabled() {
		return nil, errors.New("school service is not configured")
	}

	url := fmt.Sprintf("%s/internal/schools/by-user/%s", c.baseURL, userID.String())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("school-service returned status %d", resp.StatusCode)
	}

	var schools []model.School
	if err := json.NewDecoder(resp.Body).Decode(&schools); err != nil {
		return nil, err
	}
	return schools, nil
}

func (c *schoolClient) ListMembershipsForUser(userID uuid.UUID) ([]schoolMembership, error) {
	if !c.enabled() {
		return nil, errors.New("school service is not configured")
	}

	url := fmt.Sprintf("%s/internal/users/%s/memberships", c.baseURL, userID.String())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("school-service returned status %d", resp.StatusCode)
	}

	var rows []schoolMembership
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *schoolClient) GetMembership(schoolID, userID uuid.UUID) (*schoolMembership, error) {
	if !c.enabled() {
		return nil, errors.New("school service is not configured")
	}

	url := fmt.Sprintf("%s/internal/schools/%s/members/%s", c.baseURL, schoolID.String(), userID.String())
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("school-service returned status %d", resp.StatusCode)
	}

	var m schoolMembership
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *schoolClient) IsUserMember(schoolID, userID uuid.UUID) (bool, error) {
	_, err := c.GetMembership(schoolID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *schoolClient) ListMembersForSchool(schoolID uuid.UUID) ([]schoolMembership, error) {
	if !c.enabled() {
		return nil, errors.New("school service is not configured")
	}

	url := fmt.Sprintf("%s/internal/schools/%s/members", c.baseURL, schoolID.String())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("school-service returned status %d", resp.StatusCode)
	}

	var members []schoolMembership
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, err
	}
	return members, nil
}

func (c *schoolClient) AddMember(schoolID, userID uuid.UUID) error {
	if !c.enabled() {
		return errors.New("school service is not configured")
	}

	payload := map[string]string{
		"user_id": userID.String(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/internal/schools/%s/members", c.baseURL, schoolID.String())
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
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
		return fmt.Errorf("school-service add member returned status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (c *schoolClient) RemoveMember(schoolID, userID uuid.UUID) error {
	if !c.enabled() {
		return errors.New("school service is not configured")
	}

	url := fmt.Sprintf("%s/internal/schools/%s/members/%s", c.baseURL, schoolID.String(), userID.String())
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("school-service remove member returned status %d", resp.StatusCode)
	}
	return nil
}

// IsUserAdmin is kept for backward compatibility — checks school membership.
func (c *schoolClient) IsUserAdmin(schoolID, userID uuid.UUID) (bool, error) {
	return c.IsUserMember(schoolID, userID)
}
