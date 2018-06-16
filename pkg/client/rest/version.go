package rest

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/version"
)

type versionClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *versionClient) Show() (*version.BuildInfo, error) {
	response, err := client.httpClient.GET("/version", version.BuildInfoObject)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*version.BuildInfo), nil
}
