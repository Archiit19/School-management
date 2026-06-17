package providers

import (
	"net/http"

	"github.com/Archiit19/School-management/pkg/httpx/config"
	"github.com/Archiit19/School-management/pkg/httpx/constants"
)

// Client performs outbound service-to-service requests.
type Client interface {
	BaseURL() string
	Token() string
	URL(path string) string

	Do(req *http.Request) (*http.Response, error)
	Get(path string) (*http.Response, error)
	Post(path string, body interface{}) (*http.Response, error)
	Put(path string, body interface{}) (*http.Response, error)
	Patch(path string, body interface{}) (*http.Response, error)
	Delete(path string, body interface{}) (*http.Response, error)

	NewJSONRequest(method, path string, body interface{}) (*http.Request, error)
	DoJSON(method, path string, reqBody, respBody interface{}) (*http.Response, error)
	DoJSONExpect(method, path string, reqBody, respBody interface{}, expectStatus int) error
}

// Provider creates clients for a specific backend transport.
type Provider interface {
	Backend() constants.Backend
	New(cfg config.Config) (Client, error)
}
