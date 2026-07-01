package httpclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

type retryTransport struct {
	base http.RoundTripper
	cfg  pkgconfig.HTTPClient
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return roundTripWithRetry(t.base, req, t.cfg)
}

type statusError struct {
	status int
}

func (e *statusError) Error() string {
	return "http status " + strconv.Itoa(e.status)
}

func roundTripWithRetry(base http.RoundTripper, req *http.Request, cfg pkgconfig.HTTPClient) (*http.Response, error) {
	if err := prepareRequestBodyForRetry(req); err != nil {
		return nil, err
	}

	attempts := cfg.RetryMax + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := waitForRetry(requestContext(req), attempt, cfg); err != nil {
				return nil, err
			}
			if err := resetRequestBody(req); err != nil {
				return nil, err
			}
		}

		resp, err := base.RoundTrip(req)
		if err != nil {
			lastErr = err
			if !shouldRetry(req, nil, err, cfg) || attempt == attempts-1 {
				return nil, err
			}
			continue
		}

		if shouldRetry(req, resp, nil, cfg) && attempt < attempts-1 {
			discardResponseBody(resp)
			continue
		}

		return resp, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("httpclient: retry attempts exhausted")
}

func prepareRequestBodyForRetry(req *http.Request) error {
	if req.Body == nil || req.GetBody != nil {
		return nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	if err := req.Body.Close(); err != nil {
		return err
	}

	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	return nil
}

func resetRequestBody(req *http.Request) error {
	if req.GetBody == nil {
		return nil
	}
	body, err := req.GetBody()
	if err != nil {
		return err
	}
	req.Body = body
	return nil
}

func shouldRetry(req *http.Request, resp *http.Response, err error, cfg pkgconfig.HTTPClient) bool {
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return methodAllowsRetry(req.Method, cfg)
		}
		return methodAllowsRetry(req.Method, cfg)
	}
	if resp == nil {
		return false
	}
	switch resp.StatusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return methodAllowsRetry(req.Method, cfg)
	case http.StatusTooManyRequests:
		return methodAllowsRetry(req.Method, cfg)
	default:
		return false
	}
}

func methodAllowsRetry(method string, cfg pkgconfig.HTTPClient) bool {
	if !cfg.RetryIdempotentOnly {
		return true
	}
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func waitForRetry(ctx context.Context, attempt int, cfg pkgconfig.HTTPClient) error {
	delay := retryDelay(attempt, cfg)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func retryDelay(attempt int, cfg pkgconfig.HTTPClient) time.Duration {
	if attempt <= 0 {
		return 0
	}
	backoff := float64(cfg.RetryInitialBackoff) * math.Pow(2, float64(attempt-1))
	if max := float64(cfg.RetryMaxBackoff); max > 0 && backoff > max {
		backoff = max
	}
	delay := time.Duration(backoff)
	if cfg.RetryJitter && delay > 0 {
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))
		delay = delay/2 + jitter
	}
	return delay
}

func discardResponseBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
