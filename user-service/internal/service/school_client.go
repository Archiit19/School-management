package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/google/uuid"
)

type schoolClient struct {
	*httpclient.Client
}

func (c *schoolClient) AddMember(ctx context.Context, schoolID, userID uuid.UUID) error {
	path := fmt.Sprintf("/internal/schools/%s/members", schoolID.String())
	resp, err := c.DoJSONContext(ctx, http.MethodPost, path, map[string]string{"user_id": userID.String()}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusCreated, "school add member")
}

func (c *schoolClient) RemoveMember(ctx context.Context, schoolID, userID uuid.UUID) error {
	path := fmt.Sprintf("/internal/schools/%s/members/%s", schoolID.String(), userID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.URL(path), nil)
	if err != nil {
		return err
	}
	resp, err := c.DoContext(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "school remove member")
}

func (c *schoolClient) GetMembership(ctx context.Context, schoolID, userID uuid.UUID) error {
	path := fmt.Sprintf("/internal/schools/%s/members/%s", schoolID.String(), userID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(path), nil)
	if err != nil {
		return err
	}
	resp, err := c.DoContext(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return errors.New("not a member")
	}
	return httpclient.CheckStatus(resp, http.StatusOK, "school get member")
}

func (c *schoolClient) ListMemberUserIDs(ctx context.Context, schoolID uuid.UUID) ([]uuid.UUID, error) {
	path := fmt.Sprintf("/internal/schools/%s/members", schoolID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoContext(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := httpclient.CheckStatus(resp, http.StatusOK, "school list members"); err != nil {
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

func (c *schoolClient) ListMembershipsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	path := fmt.Sprintf("/internal/users/%s/memberships", userID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoContext(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := httpclient.CheckStatus(resp, http.StatusOK, "school list memberships"); err != nil {
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
