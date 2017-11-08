package api

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"net/http"
)

func (api *coreAPI) getUserOptional(request *http.Request) *lang.User {
	username := request.Header.Get("Username")

	if len(username) == 0 {
		return nil
	}

	return api.externalData.UserLoader.LoadUserByName(username)
}

func (api *coreAPI) getUserRequired(request *http.Request) *lang.User {
	user := api.getUserOptional(request)
	if user == nil {
		panic("Unauthorized or couldn't be loaded")
	}

	return user
}
