package builder_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Archiit19/School-management/pkg/httpx"
	"github.com/Archiit19/School-management/pkg/httpx/internal/builder"
)

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := httpx.DefaultHTTPClient()
	cfg.RetryMax = 0
	cfg.CircuitFailureThreshold = 2
	cfg.CircuitTimeout = 200 * time.Millisecond
	cfg.CircuitInterval = time.Minute

	client := httpx.NewHTTPClient("cb-test", cfg)

	for i := 0; i < 2; i++ {
		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		resp, err := client.Do(req)
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		if err == nil {
			t.Fatalf("attempt %d: expected error", i+1)
		}
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	_, err = client.Do(req)
	if err == nil {
		t.Fatal("expected circuit open error")
	}
	if !errors.Is(err, httpx.ErrCircuitOpen) {
		t.Fatalf("err = %v, want ErrCircuitOpen", err)
	}
}

func TestMethodAllowsRetry(t *testing.T) {
	cfg := httpx.DefaultHTTPClient()
	cfg.RetryIdempotentOnly = true

	if !builder.MethodAllowsRetry(http.MethodGet, cfg) {
		t.Fatal("GET should be retryable")
	}
	if builder.MethodAllowsRetry(http.MethodPost, cfg) {
		t.Fatal("POST should not be retryable when idempotent-only")
	}
}
