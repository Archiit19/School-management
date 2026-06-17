package client

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/Archiit19/School-management/pkg/httpx/constants"
	"github.com/Archiit19/School-management/pkg/httpx/internal/adapter"
	"github.com/Archiit19/School-management/pkg/httpx/internal/builder"
)

// HTTPClient is the HTTP implementation of providers.Client.
type HTTPClient struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewHTTP creates an HTTP-backed client from configuration.
func NewHTTP(cfg config.Config) *HTTPClient {
	httpCfg := cfg.HTTPConfig()
	return &HTTPClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		token:   cfg.Token,
		http:    builder.NewHTTPClient(cfg.BreakerName(), httpCfg),
	}
}

// NewHTTPWithClient creates a client with an injected *http.Client.
func NewHTTPWithClient(baseURL, token string, httpClient *http.Client) *HTTPClient {
	if httpClient == nil {
		httpClient = builder.NewHTTPClient(constants.DefaultClientName, config.LoadHTTPClientConfigFromEnv())
	}
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    httpClient,
	}
}

func (c *HTTPClient) BaseURL() string { return c.baseURL }

func (c *HTTPClient) Token() string { return c.token }

func (c *HTTPClient) URL(path string) string {
	return adapter.JoinURL(c.baseURL, path)
}

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set(constants.InternalTokenHeader, c.token)
	}
	return c.http.Do(req)
}

func (c *HTTPClient) Get(path string) (*http.Response, error) {
	return c.sendJSON(http.MethodGet, path, nil)
}

func (c *HTTPClient) Post(path string, body interface{}) (*http.Response, error) {
	return c.sendJSON(http.MethodPost, path, body)
}

func (c *HTTPClient) Put(path string, body interface{}) (*http.Response, error) {
	return c.sendJSON(http.MethodPut, path, body)
}

func (c *HTTPClient) Patch(path string, body interface{}) (*http.Response, error) {
	return c.sendJSON(http.MethodPatch, path, body)
}

func (c *HTTPClient) Delete(path string, body interface{}) (*http.Response, error) {
	return c.sendJSON(http.MethodDelete, path, body)
}

func (c *HTTPClient) sendJSON(method, path string, body interface{}) (*http.Response, error) {
	req, err := c.NewJSONRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *HTTPClient) NewJSONRequest(method, path string, body interface{}) (*http.Request, error) {
	return adapter.NewJSONRequest(c.baseURL, method, path, body)
}

func (c *HTTPClient) DoJSON(method, path string, reqBody, respBody interface{}) (*http.Response, error) {
	req, err := c.NewJSONRequest(method, path, reqBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if respBody != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			resp.Body.Close()
			return resp, err
		}
	}
	return resp, nil
}

func (c *HTTPClient) DoJSONExpect(method, path string, reqBody, respBody interface{}, expectStatus int) error {
	resp, err := c.DoJSON(method, path, reqBody, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := adapter.CheckStatus(resp, expectStatus, path); err != nil {
		return err
	}
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return err
		}
	}
	return nil
}
