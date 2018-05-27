package rest

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
)

type stateClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *stateClient) Reset(noop bool) (*api.PolicyUpdateResult, error) {
	revision, err := client.httpClient.DELETE(fmt.Sprintf("/actualstate/noop/%t", noop), api.PolicyUpdateResultObject)
	if err != nil {
		return nil, err
	}

	return revision.(*api.PolicyUpdateResult), nil
}
