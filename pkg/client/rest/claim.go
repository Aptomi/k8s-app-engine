package rest

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"

	"strings"

	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
)

type claimClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *claimClient) Status(claims []*lang.Claim, queryFlag api.ClaimQueryFlag) (*api.ClaimsStatus, error) {
	claimIds := []string{}
	for _, claim := range claims {
		claimIds = append(claimIds, claim.GetNamespace()+"^"+claim.GetName())
	}

	response, err := client.httpClient.GET(fmt.Sprintf("/policy/claim/status/%s/%s", queryFlag, strings.Join(claimIds, ",")), api.TypeClaimsStatus)
	if err != nil {
		return nil, err
	}

	if serverError, ok := response.(*api.ServerError); ok {
		return nil, fmt.Errorf("server error: %s", serverError.Error)
	}

	return response.(*api.ClaimsStatus), nil
}
