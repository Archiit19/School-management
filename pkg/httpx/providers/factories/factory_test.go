package factories_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Archiit19/School-management/pkg/httpx"
	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/Archiit19/School-management/pkg/httpx/constants"
	"github.com/Archiit19/School-management/pkg/httpx/providers/factories"
)

func TestFactoryCreatesHTTPClient(t *testing.T) {
	factory, err := factories.New(constants.BackendHTTP, &httpx.HTTPClientProvider{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := httpx.DefaultHTTPClient()
	cfg.CircuitBreakerEnabled = false

	client, err := factory.NewFromConfig(config.Config{
		BaseURL: server.URL,
		Name:    "test-service",
		HTTP:    &cfg,
	})
	if err != nil {
		t.Fatalf("NewFromConfig: %v", err)
	}

	resp, err := client.Get("/health")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestFactoryRejectsUnsupportedBackend(t *testing.T) {
	_, err := factories.New(constants.Backend("grpc"), &httpx.HTTPClientProvider{})
	if err == nil {
		t.Fatal("expected error for unsupported backend")
	}
}
