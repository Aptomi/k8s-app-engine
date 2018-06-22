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
	return &coreClient{cfg: cfg, httpClient: httpClient}
}

func (client *coreClient) Policy() client.Policy {
	return &policyClient{cfg: client.cfg, httpClient: client.httpClient}
}

func (client *coreClient) Dependency() client.Dependency {
	return &dependencyClient{cfg: client.cfg, httpClient: client.httpClient}
}

func (client *coreClient) Revision() client.Revision {
	return &revisionClient{cfg: client.cfg, httpClient: client.httpClient}
}

func (client *coreClient) State() client.State {
	return &stateClient{cfg: client.cfg, httpClient: client.httpClient}
}

func (client *coreClient) User() client.User {
	return &userClient{cfg: client.cfg, httpClient: client.httpClient}
}

func (client *coreClient) Version() client.Version {
	return &versionClient{cfg: client.cfg, httpClient: client.httpClient}
}
