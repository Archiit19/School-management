package config

import (
	"testing"
	"time"
)

func TestDefaultHTTPClient(t *testing.T) {
	cfg := DefaultHTTPClient()
	if cfg.Timeout != 10*time.Second {
		t.Fatalf("Timeout = %v, want 10s", cfg.Timeout)
	}
	if cfg.RetryMax != 3 {
		t.Fatalf("RetryMax = %d, want 3", cfg.RetryMax)
	}
	if !cfg.CircuitBreakerEnabled {
		t.Fatal("CircuitBreakerEnabled should default to true")
	}
}
