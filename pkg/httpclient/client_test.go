package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
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

	cfg := pkgconfig.DefaultHTTPClient()
	cfg.CircuitBreakerEnabled = false

	client := NewFromConfig(ClientConfig{
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
