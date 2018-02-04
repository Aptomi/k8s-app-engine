package rest

import (
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine"
)

type stateClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *stateClient) Reset() (*engine.Revision, error) {
	revision, err := client.httpClient.DELETE("/actualstate", engine.RevisionObject)
	if err != nil {
		return nil, err
	}

	return revision.(*engine.Revision), nil
}
