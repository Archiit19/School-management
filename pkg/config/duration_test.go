package config

import (
	"testing"
	"time"
)

func TestGetEnvDuration(t *testing.T) {
	t.Setenv("TEST_HTTP_DURATION", "250ms")
	if got := GetEnvDuration("TEST_HTTP_DURATION", time.Second); got != 250*time.Millisecond {
		t.Fatalf("got %v, want 250ms", got)
	}
	if got := GetEnvDuration("TEST_HTTP_DURATION_MISSING", time.Second); got != time.Second {
		t.Fatalf("fallback got %v, want 1s", got)
	}
}
