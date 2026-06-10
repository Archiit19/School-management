package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/google/uuid"
)

type academicClient struct {
	*httpclient.Client
}

func newAcademicClient(cfg *config.Config, client *http.Client) *academicClient {
	if client == nil {
		return &academicClient{Client: httpclient.New(cfg.AcademicServiceURL, cfg.InternalServiceToken)}
	}
	return &academicClient{Client: httpclient.NewWithHTTP(cfg.AcademicServiceURL, cfg.InternalServiceToken, client)}
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
	resp, err := c.DoJSON(http.MethodPost, "/internal/enrollments", body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatus(resp, http.StatusOK, "academic upsert enrollment")
}

func (c *academicClient) GetEnrollment(userID, schoolID uuid.UUID) (*studentEnrollment, error) {
	path := fmt.Sprintf("/internal/enrollments/%s?school_id=%s", userID.String(), url.QueryEscape(schoolID.String()))
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
		return nil, nil
	}
	if err := httpclient.CheckStatus(resp, http.StatusOK, "academic get enrollment"); err != nil {
		return nil, err
	}
	var row studentEnrollment
	if err := json.NewDecoder(resp.Body).Decode(&row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (c *academicClient) DeleteEnrollment(userID, schoolID uuid.UUID) error {
	path := fmt.Sprintf("/internal/enrollments/%s?school_id=%s", userID.String(), url.QueryEscape(schoolID.String()))
	req, err := http.NewRequest(http.MethodDelete, c.URL(path), nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return httpclient.CheckStatusAny(resp, "academic delete enrollment", http.StatusOK, http.StatusNotFound)
}
