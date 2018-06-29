package rest

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type policyClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *policyClient) Show(gen runtime.Generation) (*engine.PolicyData, error) {
	response, err := client.httpClient.GET(fmt.Sprintf("/policy/gen/%d", gen), engine.TypePolicyData)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*engine.PolicyData), nil
}

func (client *policyClient) Apply(updated []runtime.Object, noop bool, logLevel logrus.Level) (*api.PolicyUpdateResult, error) {
	response, err := client.httpClient.POSTSlice(fmt.Sprintf("/policy/noop/%t/loglevel/%s", noop, logLevel.String()), api.TypePolicyUpdateResult, updated)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*api.PolicyUpdateResult), nil
}

func (client *policyClient) Delete(updated []runtime.Object, noop bool, logLevel logrus.Level) (*api.PolicyUpdateResult, error) {
	response, err := client.httpClient.DELETESlice(fmt.Sprintf("/policy/noop/%t/loglevel/%s", noop, logLevel.String()), api.TypePolicyUpdateResult, updated)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*api.PolicyUpdateResult), nil
}
