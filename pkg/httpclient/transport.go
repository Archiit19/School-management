package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

// OutboundHTTP returns a production *http.Client without internal auth (e.g. JWT forwarding).
func OutboundHTTP(name string) *http.Client {
	return NewHTTPClient(name, pkgconfig.LoadHTTPClientConfigFromEnv())
}

// NewHTTPClient builds a production-ready *http.Client with pooling, retry, and optional circuit breaker.
func NewHTTPClient(name string, cfg pkgconfig.HTTPClient) *http.Client {
	base := newBaseTransport(cfg)
	var transport http.RoundTripper = base
	if cfg.CircuitBreakerEnabled {
		transport = newCircuitBreakerTransport(name, transport, cfg)
	} else {
		transport = &retryTransport{base: transport, cfg: cfg}
	}
	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}
}

func newBaseTransport(cfg pkgconfig.HTTPClient) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   cfg.DialTimeout,
		KeepAlive: 30 * time.Second,
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.MaxConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// requestContext returns the request context or background when unset.
func requestContext(req *http.Request) context.Context {
	if req.Context() != nil {
		return req.Context()
	}
	return context.Background()
}
