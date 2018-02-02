package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/julienschmidt/httprouter"
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
		panic("Unauthorized or user couldn't be loaded")
	}

	return user
}

func (api *coreAPI) authenticateUser(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	username := request.PostFormValue("username")
	password := request.PostFormValue("password")
	_, err := api.externalData.UserLoader.Authenticate(username, password)
	if err != nil {
		serverErr := NewServerError(fmt.Sprintf("Authentication error: %s", err))
		api.contentType.WriteOneWithStatus(writer, request, serverErr, http.StatusUnauthorized)
	} else {
		api.contentType.WriteOneWithStatus(writer, request, nil, http.StatusOK)
	}
}
