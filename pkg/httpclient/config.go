package httpclient

import (
	"strings"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

// ClientConfig configures an internal service HTTP client.
type ClientConfig struct {
	BaseURL string
	Token   string
	// Name identifies the downstream dependency for circuit-breaker metrics/logs.
	// When empty, it is derived from BaseURL.
	Name string
	// HTTP overrides outbound transport/retry/circuit settings.
	// When nil, settings are loaded from environment via pkg/config.LoadHTTPClientConfigFromEnv.
	HTTP *pkgconfig.HTTPClient
}

func (c ClientConfig) httpConfig() pkgconfig.HTTPClient {
	if c.HTTP != nil {
		return *c.HTTP
	}
	return pkgconfig.LoadHTTPClientConfigFromEnv()
}

func (c ClientConfig) breakerName() string {
	if name := strings.TrimSpace(c.Name); name != "" {
		return name
	}
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		return "httpclient"
	}
	return base
}
