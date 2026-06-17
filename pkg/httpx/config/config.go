package config

import (
	"strings"

	"github.com/Archiit19/School-management/pkg/httpx/constants"
)

// Config configures an outbound service client.
type Config struct {
	Backend constants.Backend
	BaseURL string
	Token   string
	// Name identifies the downstream dependency for circuit-breaker metrics/logs.
	Name string
	// HTTP overrides outbound transport/retry/circuit settings.
	HTTP *HTTPClient
}

// HTTPConfig returns transport settings, loading from environment when unset.
func (c Config) HTTPConfig() HTTPClient {
	if c.HTTP != nil {
		return *c.HTTP
	}
	return LoadHTTPClientConfigFromEnv()
}

// BreakerName returns the circuit-breaker name for this client.
func (c Config) BreakerName() string {
	if name := strings.TrimSpace(c.Name); name != "" {
		return name
	}
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		return constants.DefaultClientName
	}
	return base
}
