package config

import (
	"time"

	pkgconfig "github.com/Archiit19/School-management/pkg/config"
)

// HTTPClient holds transport, retry, and circuit-breaker settings for outbound calls.
type HTTPClient struct {
	Timeout               time.Duration
	DialTimeout           time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	IdleConnTimeout       time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	MaxConnsPerHost       int

	RetryMax            int
	RetryInitialBackoff time.Duration
	RetryMaxBackoff     time.Duration
	RetryJitter         bool
	RetryIdempotentOnly bool

	CircuitBreakerEnabled   bool
	CircuitMaxRequests      uint32
	CircuitInterval         time.Duration
	CircuitTimeout          time.Duration
	CircuitFailureThreshold uint32
}

// DefaultHTTPClient returns production-oriented defaults for internal service calls.
func DefaultHTTPClient() HTTPClient {
	return HTTPClient{
		Timeout:               10 * time.Second,
		DialTimeout:           2 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		MaxConnsPerHost:       0,

		RetryMax:            3,
		RetryInitialBackoff: 100 * time.Millisecond,
		RetryMaxBackoff:     2 * time.Second,
		RetryJitter:         true,
		RetryIdempotentOnly: true,

		CircuitBreakerEnabled:   true,
		CircuitMaxRequests:      1,
		CircuitInterval:         60 * time.Second,
		CircuitTimeout:          30 * time.Second,
		CircuitFailureThreshold: 5,
	}
}

// LoadHTTPClientConfigFromEnv loads HTTP client settings from environment variables.
func LoadHTTPClientConfigFromEnv() HTTPClient {
	def := DefaultHTTPClient()
	return HTTPClient{
		Timeout:               pkgconfig.GetEnvDuration("HTTP_CLIENT_TIMEOUT", def.Timeout),
		DialTimeout:           pkgconfig.GetEnvDuration("HTTP_DIAL_TIMEOUT", def.DialTimeout),
		TLSHandshakeTimeout:   pkgconfig.GetEnvDuration("HTTP_TLS_HANDSHAKE_TIMEOUT", def.TLSHandshakeTimeout),
		ResponseHeaderTimeout: pkgconfig.GetEnvDuration("HTTP_RESPONSE_HEADER_TIMEOUT", def.ResponseHeaderTimeout),
		IdleConnTimeout:       pkgconfig.GetEnvDuration("HTTP_IDLE_CONN_TIMEOUT", def.IdleConnTimeout),
		MaxIdleConns:          pkgconfig.GetEnvInt("HTTP_MAX_IDLE_CONNS", def.MaxIdleConns),
		MaxIdleConnsPerHost:   pkgconfig.GetEnvInt("HTTP_MAX_IDLE_CONNS_PER_HOST", def.MaxIdleConnsPerHost),
		MaxConnsPerHost:       pkgconfig.GetEnvInt("HTTP_MAX_CONNS_PER_HOST", def.MaxConnsPerHost),

		RetryMax:            pkgconfig.GetEnvInt("HTTP_RETRY_MAX", def.RetryMax),
		RetryInitialBackoff: pkgconfig.GetEnvDuration("HTTP_RETRY_INITIAL_BACKOFF", def.RetryInitialBackoff),
		RetryMaxBackoff:     pkgconfig.GetEnvDuration("HTTP_RETRY_MAX_BACKOFF", def.RetryMaxBackoff),
		RetryJitter:         pkgconfig.GetEnvBool("HTTP_RETRY_JITTER", def.RetryJitter),
		RetryIdempotentOnly: pkgconfig.GetEnvBool("HTTP_RETRY_IDEMPOTENT_ONLY", def.RetryIdempotentOnly),

		CircuitBreakerEnabled:   pkgconfig.GetEnvBool("HTTP_CIRCUIT_BREAKER_ENABLED", def.CircuitBreakerEnabled),
		CircuitMaxRequests:      uint32(pkgconfig.GetEnvInt("HTTP_CIRCUIT_MAX_REQUESTS", int(def.CircuitMaxRequests))),
		CircuitInterval:         pkgconfig.GetEnvDuration("HTTP_CIRCUIT_INTERVAL", def.CircuitInterval),
		CircuitTimeout:          pkgconfig.GetEnvDuration("HTTP_CIRCUIT_TIMEOUT", def.CircuitTimeout),
		CircuitFailureThreshold: uint32(pkgconfig.GetEnvInt("HTTP_CIRCUIT_FAILURE_THRESHOLD", int(def.CircuitFailureThreshold))),
	}
}
