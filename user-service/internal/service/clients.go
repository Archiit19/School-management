package service

import (
	"github.com/Archiit19/School-management/pkg/httpx"
	"github.com/Archiit19/School-management/user-service/internal/config"
)

func outboundHTTPConfig() *httpx.HTTPClient {
	cfg := httpx.LoadHTTPClientConfigFromEnv()
	return &cfg
}

func newOutboundClient(baseURL, token, name string, httpCfg *httpx.HTTPClient) httpx.Client {
	return httpx.NewFromConfig(httpx.ClientConfig{
		BaseURL: baseURL,
		Token:   token,
		Name:    name,
		HTTP:    httpCfg,
	})
}

func newServiceClients(cfg *config.Config, httpCfg *httpx.HTTPClient) (*authClient, *schoolClient, *academicClient) {
	return &authClient{Client: newOutboundClient(cfg.AuthServiceURL, cfg.InternalServiceToken, "auth-service", httpCfg)},
		&schoolClient{Client: newOutboundClient(cfg.SchoolServiceURL, cfg.InternalServiceToken, "school-service", httpCfg)},
		&academicClient{Client: newOutboundClient(cfg.AcademicServiceURL, cfg.InternalServiceToken, "academic-service", httpCfg)}
}
