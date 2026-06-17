package constants

import "time"

const (
	InternalTokenHeader = "X-Internal-Token"
	DefaultClientName   = "httpx"
)

const DefaultTimeout = 8 * time.Second

// Backend identifies the transport used for outbound service calls.
type Backend string

const (
	BackendHTTP Backend = "http"
)
