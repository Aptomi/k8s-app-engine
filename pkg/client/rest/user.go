package rest

import (
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
)

type userClient struct {
	cfg        *config.Client
	httpClient http.Client
}

func (client *userClient) Login(username, password string) (*api.AuthSuccess, error) {
	authReq := &api.AuthRequest{
		TypeKind: api.TypeAuthRequest.GetTypeKind(),
		Username: username,
		Password: password,
	}
	authSuccess, err := client.httpClient.POST("/user/login", api.TypeAuthSuccess, authReq)
	if err != nil {
		return nil, err
	}

	return authSuccess.(*api.AuthSuccess), nil
}
