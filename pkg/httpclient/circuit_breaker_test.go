package httpclient

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := pkgconfig.DefaultHTTPClient()
	cfg.RetryMax = 0
	cfg.CircuitFailureThreshold = 2
	cfg.CircuitTimeout = 200 * time.Millisecond
	cfg.CircuitInterval = time.Minute

	client := NewHTTPClient("cb-test", cfg)

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
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("err = %v, want ErrCircuitOpen", err)
	}
}
