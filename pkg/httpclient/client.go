package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/pkg/correlation"
)

const InternalTokenHeader = "X-Internal-Token"

const DefaultTimeout = 8 * time.Second

// Client is a base HTTP client for internal service-to-service calls.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// New creates an internal HTTP client using environment-based production defaults.
func New(baseURL, token string) *Client {
	return NewFromConfig(ClientConfig{BaseURL: baseURL, Token: token})
}

// NewFromConfig creates an internal HTTP client with tuned transport, retry, and circuit breaker.
func NewFromConfig(cfg ClientConfig) *Client {
	httpCfg := cfg.httpConfig()
	return &Client{
		BaseURL: strings.TrimRight(cfg.BaseURL, "/"),
		Token:   cfg.Token,
		HTTP:    NewHTTPClient(cfg.breakerName(), httpCfg),
	}
}

// NewWithHTTP allows injecting a custom http.Client (e.g. tests or shared transport).
func NewWithHTTP(baseURL, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = NewHTTPClient("httpclient", pkgconfig.LoadHTTPClientConfigFromEnv())
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTP:    httpClient,
	}
}

// URL joins the base URL with a path.
func (c *Client) URL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.BaseURL + path
}

// Do sends a request with the internal token header when configured.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.applyDefaultHeaders(req)
	return c.HTTP.Do(req)
}

// DoContext is like Do but attaches ctx to the request for cancellation and deadlines.
func (c *Client) DoContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.WithContext(ctx)
	c.applyDefaultHeaders(req)
	return c.HTTP.Do(req)
}

func (c *Client) applyDefaultHeaders(req *http.Request) {
	if c.Token != "" {
		req.Header.Set(InternalTokenHeader, c.Token)
	}
	if req.Header.Get(correlation.RequestIDHeader) == "" {
		if id := correlation.RequestIDFromContext(req.Context()); id != "" {
			req.Header.Set(correlation.RequestIDHeader, id)
		}
	}
}

// NewJSONRequest builds a request with JSON content type when body is non-nil.
func (c *Client) NewJSONRequest(method, path string, body interface{}) (*http.Request, error) {
	return c.NewJSONRequestContext(context.Background(), method, path, body)
}

// NewJSONRequestContext builds a JSON request bound to ctx.
func (c *Client) NewJSONRequestContext(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.URL(path), reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// ReadErrorBody reads response body for error messages.
func ReadErrorBody(resp *http.Response) string {
	b, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(b))
}

// CheckStatus returns an error when status does not match expectStatus.
func CheckStatus(resp *http.Response, expectStatus int, context string) error {
	if resp.StatusCode == expectStatus {
		return nil
	}
	msg := ReadErrorBody(resp)
	if msg != "" {
		return fmt.Errorf("%s status %d: %s", context, resp.StatusCode, msg)
	}
	return fmt.Errorf("%s status %d", context, resp.StatusCode)
}

// CheckStatusAny returns an error when status is not one of the allowed codes.
func CheckStatusAny(resp *http.Response, context string, allowed ...int) error {
	for _, code := range allowed {
		if resp.StatusCode == code {
			return nil
		}
	}
	msg := ReadErrorBody(resp)
	if msg != "" {
		return fmt.Errorf("%s status %d: %s", context, resp.StatusCode, msg)
	}
	return fmt.Errorf("%s status %d", context, resp.StatusCode)
}

// DoJSON sends a JSON request and decodes the response when respBody is non-nil.
func (c *Client) DoJSON(method, path string, reqBody, respBody interface{}) (*http.Response, error) {
	return c.DoJSONContext(context.Background(), method, path, reqBody, respBody)
}

// DoJSONContext sends a JSON request bound to ctx.
func (c *Client) DoJSONContext(ctx context.Context, method, path string, reqBody, respBody interface{}) (*http.Response, error) {
	req, err := c.NewJSONRequestContext(ctx, method, path, reqBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoContext(ctx, req)
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

// DoJSONExpect sends JSON and checks for an expected status before decoding.
func (c *Client) DoJSONExpect(method, path string, reqBody, respBody interface{}, expectStatus int) error {
	return c.DoJSONExpectContext(context.Background(), method, path, reqBody, respBody, expectStatus)
}

// DoJSONExpectContext is DoJSONExpect with request context.
func (c *Client) DoJSONExpectContext(ctx context.Context, method, path string, reqBody, respBody interface{}, expectStatus int) error {
	resp, err := c.DoJSONContext(ctx, method, path, reqBody, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := CheckStatus(resp, expectStatus, path); err != nil {
		return err
	}
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return err
		}
	}
	return nil
}
