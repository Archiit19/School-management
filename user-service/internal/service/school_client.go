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

	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/google/uuid"
)

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

func (c *schoolClient) do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("X-Internal-Token", c.token)
	}
	return c.client.Do(req)
}

func (c *schoolClient) AddMember(schoolID, userID uuid.UUID) error {
	payload := map[string]string{"user_id": userID.String()}
	body, _ := json.Marshal(payload)
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
		return fmt.Errorf("school add member status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (c *schoolClient) RemoveMember(schoolID, userID uuid.UUID) error {
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
		return fmt.Errorf("school remove member status %d", resp.StatusCode)
	}
	return nil
}

func (c *schoolClient) GetMembership(schoolID, userID uuid.UUID) error {
	url := fmt.Sprintf("%s/internal/schools/%s/members/%s", c.baseURL, schoolID.String(), userID.String())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return errors.New("not a member")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("school get member status %d", resp.StatusCode)
	}
	return nil
}

func (c *schoolClient) ListMemberUserIDs(schoolID uuid.UUID) ([]uuid.UUID, error) {
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
		return nil, fmt.Errorf("school list members status %d", resp.StatusCode)
	}
	var rows []struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.UserID
	}
	return ids, nil
}

func (c *schoolClient) ListMembershipsForUser(userID uuid.UUID) ([]uuid.UUID, error) {
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
		return nil, fmt.Errorf("school list memberships status %d", resp.StatusCode)
	}
	var rows []struct {
		SchoolID uuid.UUID `json:"school_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.SchoolID
	}
	return ids, nil
}
