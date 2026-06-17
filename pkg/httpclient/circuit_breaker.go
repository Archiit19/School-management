package httpclient

import (
	"errors"
	"fmt"
	"net/http"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/sony/gobreaker"
)

// ErrCircuitOpen is returned when the circuit breaker is open for a dependency.
var ErrCircuitOpen = errors.New("httpclient: circuit breaker open")

func newCircuitBreakerTransport(name string, base http.RoundTripper, cfg pkgconfig.HTTPClient) http.RoundTripper {
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
	cfg  pkgconfig.HTTPClient
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
		return nil, errors.New("httpclient: unexpected circuit breaker result")
	}
	return resp, nil
}
