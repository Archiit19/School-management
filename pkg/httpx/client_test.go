package httpx

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Archiit19/School-management/pkg/httpx/config"
)

func TestNewFromConfigSetsInternalToken(t *testing.T) {
	const token = "test-internal-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(InternalTokenHeader); got != token {
			t.Fatalf("token = %q, want %q", got, token)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := DefaultHTTPClient()
	cfg.CircuitBreakerEnabled = false

	client := NewFromConfig(config.Config{
		BaseURL: server.URL,
		Token:   token,
		Name:    "test-service",
		HTTP:    &cfg,
	})

	req, err := http.NewRequest(http.MethodGet, client.URL("/health"), nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCheckStatus(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewReader([]byte("bad input"))),
	}
	if err := CheckStatus(resp, http.StatusOK, "test"); err == nil {
		t.Fatal("expected error")
	}
}

func TestHTTPClientVerbs(t *testing.T) {
	const token = "test-internal-token"
	var gotMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if got := r.Header.Get(InternalTokenHeader); got != token {
			t.Fatalf("token = %q, want %q", got, token)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewFromConfig(config.Config{
		BaseURL: server.URL,
		Token:   token,
		Name:    "test-service",
		HTTP:    httpTestConfig(),
	})

	tests := []struct {
		name   string
		method string
		call   func() (*http.Response, error)
	}{
		{"get", http.MethodGet, func() (*http.Response, error) { return client.Get("/health") }},
		{"post", http.MethodPost, func() (*http.Response, error) { return client.Post("/health", map[string]string{"ok": "1"}) }},
		{"put", http.MethodPut, func() (*http.Response, error) { return client.Put("/health", map[string]string{"ok": "1"}) }},
		{"patch", http.MethodPatch, func() (*http.Response, error) { return client.Patch("/health", map[string]string{"ok": "1"}) }},
		{"delete", http.MethodDelete, func() (*http.Response, error) { return client.Delete("/health", nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.call()
			if err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
			defer resp.Body.Close()
			_, _ = io.Copy(io.Discard, resp.Body)

			if gotMethod != tt.method {
				t.Fatalf("method = %q, want %q", gotMethod, tt.method)
			}
		})
	}
}

func httpTestConfig() *HTTPClient {
	cfg := DefaultHTTPClient()
	cfg.CircuitBreakerEnabled = false
	return &cfg
}
