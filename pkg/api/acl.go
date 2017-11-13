package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type userRolesWrapper struct {
	Data interface{}
}

func (g *userRolesWrapper) GetKind() string {
	return "userRoles"
}

func (api *coreAPI) handleUserRoles(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	policy, _, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while getting policy: %s", err))
	}

	systemNamespace := policy.Namespace[runtime.SystemNS]
	var aclResolver *lang.ACLResolver
	if systemNamespace != nil {
		aclResolver = lang.NewACLResolver(systemNamespace.ACLRules)
	} else {
		aclResolver = lang.NewACLResolver(lang.NewGlobalRules())
	}

	data := make(map[string]interface{})
	users := api.externalData.UserLoader.LoadUsersAll().Users
	for _, user := range users {
		roleMap, errRoleMap := aclResolver.GetUserRoleMap(user)
		if errRoleMap != nil {
			panic(fmt.Sprintf("error while retrieving user role map for '%s': %s", user.Name, err))
		}
		data[user.Name] = roleMap
	}
	api.contentType.Write(writer, request, &userRolesWrapper{Data: data})
}
