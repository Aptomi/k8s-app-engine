package rest

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

type policyClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *policyClient) Show(gen runtime.Generation) (*engine.PolicyData, error) {
	response, err := client.httpClient.GET(fmt.Sprintf("/policy/gen/%d", gen), engine.PolicyDataObject)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*engine.PolicyData), nil
}

func (client *policyClient) Apply(updated []runtime.Object) (*api.PolicyUpdateResult, error) {
	response, err := client.httpClient.POSTSlice("/policy", api.PolicyUpdateResultObject, updated)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*api.PolicyUpdateResult), nil
}

func (client *policyClient) Delete(deleted []string) (*api.PolicyUpdateResult, error) {
	// todo(slukjanov): implement policy delete handling
	panic("implement me")
}
