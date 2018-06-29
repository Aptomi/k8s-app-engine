package rest

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

type stateClient struct {
	cfg        *config.Client
	httpClient http.Client
}

type stateEnforceObj struct {
}

func (stateEnforce *stateEnforceObj) GetKind() runtime.Kind {
	return "state-enforce-obj"
}

func (client *stateClient) Reset(noop bool) (*api.PolicyUpdateResult, error) {
	revision, err := client.httpClient.POST(fmt.Sprintf("/state/enforce/noop/%t", noop), api.TypePolicyUpdateResult, &stateEnforceObj{})
	if err != nil {
		return nil, err
	}

	return revision.(*api.PolicyUpdateResult), nil
}
