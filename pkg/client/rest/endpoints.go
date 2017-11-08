package rest

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"

	"github.com/Aptomi/aptomi/pkg/config"
)

type endpointsClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *endpointsClient) Show() (*api.Endpoints, error) {
	response, err := client.httpClient.GET("/endpoints", api.EndpointsObject)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*api.Endpoints), nil
}
