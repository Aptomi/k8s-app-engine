package rest

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"

	"strings"

	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
)

type dependencyClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *dependencyClient) Status(dependencies []*lang.Dependency, queryFlag api.DependencyQueryFlag) (*api.DependenciesStatus, error) {
	dependencyIds := []string{}
	for _, d := range dependencies {
		dependencyIds = append(dependencyIds, d.GetNamespace()+"^"+d.GetName())
	}

	response, err := client.httpClient.GET(fmt.Sprintf("/policy/dependency/status/%s/%s", queryFlag, strings.Join(dependencyIds, ",")), api.DependenciesStatusObject)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*api.DependenciesStatus), nil
}
