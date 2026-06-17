package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

func TestRoundTripWithRetryRetriesGETOn503(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer server.Close()

	cfg := pkgconfig.DefaultHTTPClient()
	cfg.RetryMax = 2
	cfg.RetryInitialBackoff = 1 * time.Millisecond
	cfg.RetryMaxBackoff = 2 * time.Millisecond
	cfg.RetryJitter = false

	client := NewHTTPClient("retry-test", cfg)
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestRoundTripWithRetryDoesNotRetryPOST(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cfg := pkgconfig.DefaultHTTPClient()
	cfg.RetryMax = 3
	cfg.RetryIdempotentOnly = true
	cfg.CircuitBreakerEnabled = false

	client := NewHTTPClient("retry-test", cfg)
	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader([]byte("payload")))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestMethodAllowsRetry(t *testing.T) {
	cfg := pkgconfig.DefaultHTTPClient()
	cfg.RetryIdempotentOnly = true

	if !methodAllowsRetry(http.MethodGet, cfg) {
		t.Fatal("GET should be retryable")
	}
	if methodAllowsRetry(http.MethodPost, cfg) {
		t.Fatal("POST should not be retryable when idempotent-only")
	}
}
