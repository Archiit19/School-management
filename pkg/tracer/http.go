package tracer

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// WrapTransport instruments outbound HTTP requests with client spans and trace propagation.
// When tracing is disabled, the base transport is returned unchanged.
func WrapTransport(base http.RoundTripper) http.RoundTripper {
	return WrapTransportWithEnabled(base, Enabled())
}

// WrapTransportWithEnabled instruments outbound HTTP when enabled is true.
func WrapTransportWithEnabled(base http.RoundTripper, enabled bool) http.RoundTripper {
	if !enabled || base == nil {
		return base
	}
	return otelhttp.NewTransport(base)
}
