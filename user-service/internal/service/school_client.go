package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Archiit19/School-management/pkg/httpx"
	"github.com/google/uuid"
)

type schoolClient struct {
	httpx.Client
}

func (c *schoolClient) AddMember(schoolID, userID uuid.UUID) error {
	path := fmt.Sprintf("/internal/schools/%s/members", schoolID.String())
	resp, err := c.DoJSON(http.MethodPost, path, map[string]string{"user_id": userID.String()}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpx.CheckStatus(resp, http.StatusCreated, "school add member")
}

func (c *schoolClient) RemoveMember(schoolID, userID uuid.UUID) error {
	path := fmt.Sprintf("/internal/schools/%s/members/%s", schoolID.String(), userID.String())
	req, err := http.NewRequest(http.MethodDelete, c.URL(path), nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpx.CheckStatus(resp, http.StatusOK, "school remove member")
}

func (c *schoolClient) GetMembership(schoolID, userID uuid.UUID) error {
	path := fmt.Sprintf("/internal/schools/%s/members/%s", schoolID.String(), userID.String())
	req, err := http.NewRequest(http.MethodGet, c.URL(path), nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return errors.New("not a member")
	}
	return httpx.CheckStatus(resp, http.StatusOK, "school get member")
}

func (c *schoolClient) ListMemberUserIDs(schoolID uuid.UUID) ([]uuid.UUID, error) {
	path := fmt.Sprintf("/internal/schools/%s/members", schoolID.String())
	req, err := http.NewRequest(http.MethodGet, c.URL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := httpx.CheckStatus(resp, http.StatusOK, "school list members"); err != nil {
		return nil, err
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
	path := fmt.Sprintf("/internal/users/%s/memberships", userID.String())
	req, err := http.NewRequest(http.MethodGet, c.URL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := httpx.CheckStatus(resp, http.StatusOK, "school list memberships"); err != nil {
		return nil, err
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
