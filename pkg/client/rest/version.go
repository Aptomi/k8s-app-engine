package rest

import (
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
)

type versionClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *versionClient) Show() (*api.Version, error) {
	// todo(slukjanov): implement version show
	panic("implement me")
}
