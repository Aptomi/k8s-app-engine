package rest

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"

	"github.com/Aptomi/aptomi/pkg/config"
)

type revisionClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *revisionClient) Show(gen runtime.Generation) (*engine.Revision, error) {
	response, err := client.httpClient.GET(fmt.Sprintf("/revision/gen/%d", gen), engine.TypeRevision)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*engine.Revision), nil
}
