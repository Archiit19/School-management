package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/google/uuid"
)

type academicClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func newAcademicClient(cfg *config.Config, client *http.Client) *academicClient {
	return &academicClient{
		baseURL: strings.TrimRight(cfg.AcademicServiceURL, "/"),
		token:   cfg.InternalServiceToken,
		client:  client,
	}
}

type studentEnrollment struct {
	UserID    string  `json:"user_id"`
	ClassID   string  `json:"class_id"`
	SectionID *string `json:"section_id,omitempty"`
}

func (c *academicClient) UpsertEnrollment(userID, schoolID uuid.UUID, classID string, sectionID string) error {
	body := map[string]string{
		"user_id":   userID.String(),
		"school_id": schoolID.String(),
		"class_id":  classID,
	}
	if strings.TrimSpace(sectionID) != "" {
		body["section_id"] = sectionID
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/enrollments", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("academic upsert enrollment status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (c *academicClient) GetEnrollment(userID, schoolID uuid.UUID) (*studentEnrollment, error) {
	u := fmt.Sprintf("%s/internal/enrollments/%s?school_id=%s", c.baseURL, userID.String(), url.QueryEscape(schoolID.String()))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("academic get enrollment status %d", resp.StatusCode)
	}
	var row studentEnrollment
	if err := json.NewDecoder(resp.Body).Decode(&row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (c *academicClient) DeleteEnrollment(userID, schoolID uuid.UUID) error {
	u := fmt.Sprintf("%s/internal/enrollments/%s?school_id=%s", c.baseURL, userID.String(), url.QueryEscape(schoolID.String()))
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Token", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("academic delete enrollment status %d", resp.StatusCode)
	}
	return nil
}
