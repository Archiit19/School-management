package builder

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/sony/gobreaker"
)

// ErrCircuitOpen is returned when the circuit breaker is open for a dependency.
var ErrCircuitOpen = errors.New("httpx: circuit breaker open")

// NewHTTPClient builds a production-ready *http.Client with pooling, retry, and optional circuit breaker.
func NewHTTPClient(name string, cfg config.HTTPClient) *http.Client {
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

func newBaseTransport(cfg config.HTTPClient) *http.Transport {
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

func requestContext(req *http.Request) context.Context {
	if req.Context() != nil {
		return req.Context()
	}
	return context.Background()
}

func newCircuitBreakerTransport(name string, base http.RoundTripper, cfg config.HTTPClient) http.RoundTripper {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: cfg.CircuitMaxRequests,
		Interval:    cfg.CircuitInterval,
		Timeout:     cfg.CircuitTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.CircuitFailureThreshold
		},
	}
	cb := gobreaker.NewCircuitBreaker(settings)
	return &circuitBreakerRoundTripper{
		base: base,
		cfg:  cfg,
		cb:   cb,
	}
}

type circuitBreakerRoundTripper struct {
	base http.RoundTripper
	cfg  config.HTTPClient
	cb   *gobreaker.CircuitBreaker
}

func (t *circuitBreakerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	result, err := t.cb.Execute(func() (interface{}, error) {
		resp, err := roundTripWithRetry(t.base, req, t.cfg)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= http.StatusInternalServerError {
			status := resp.StatusCode
			discardResponseBody(resp)
			return nil, &statusError{status: status}
		}
		return resp, nil
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, fmt.Errorf("%w: %s", ErrCircuitOpen, t.cb.Name())
		}
		return nil, err
	}
	resp, ok := result.(*http.Response)
	if !ok {
		return nil, errors.New("httpx: unexpected circuit breaker result")
	}
	return resp, nil
}
