package service

import (
	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/Archiit19/School-management/user-service/internal/config"
)

func outboundHTTPConfig() *pkgconfig.HTTPClient {
	cfg := pkgconfig.LoadHTTPClientConfigFromEnv()
	return &cfg
}

func newOutboundClient(baseURL, token, name string, httpCfg *pkgconfig.HTTPClient) *httpclient.Client {
	return httpclient.NewFromConfig(httpclient.ClientConfig{
		BaseURL: baseURL,
		Token:   token,
		Name:    name,
		HTTP:    httpCfg,
	})
}

func newServiceClients(cfg *config.Config, httpCfg *pkgconfig.HTTPClient) (*authClient, *schoolClient, *academicClient) {
	return &authClient{Client: newOutboundClient(cfg.AuthServiceURL, cfg.InternalServiceToken, "auth-service", httpCfg)},
		&schoolClient{Client: newOutboundClient(cfg.SchoolServiceURL, cfg.InternalServiceToken, "school-service", httpCfg)},
		&academicClient{Client: newOutboundClient(cfg.AcademicServiceURL, cfg.InternalServiceToken, "academic-service", httpCfg)}
}
