package rest

import (
	"github.com/Aptomi/aptomi/pkg/client"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
)

type coreClient struct {
	cfg        *config.Client
	httpClient http.Client
}

// New returns new instance of the Core API client http rest implementation
func New(cfg *config.Client, httpClient http.Client) client.Core {
	return &coreClient{cfg, httpClient}
}

func (client *coreClient) Policy() client.Policy {
	return &policyClient{client.cfg, client.httpClient}
}

func (client *coreClient) Endpoints() client.Endpoints {
	return &endpointsClient{client.cfg, client.httpClient}
}

func (client *coreClient) Version() client.Version {
	return &versionClient{client.cfg, client.httpClient}
}
