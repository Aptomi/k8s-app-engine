package visibility

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin/deployment"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"sort"
)

type rLink struct {
	Name string
	Link string
}

type rEndpoint struct {
	Service         string
	Context         string
	ContextWithKeys string
	Component       string
	Links           []rLink
}

type userEndpoints struct {
	User      *User
	Endpoints []rEndpoint
}

type endpointsView struct {
	Endpoints []userEndpoints
}

// Endpoints returns a view with all endpoints
func Endpoints(currentUserID string) endpointsView {
	state := resolve.LoadResolvedState()
	users := state.UserLoader.LoadUsersAll().Users

	uR := endpointsView{make([]userEndpoints, 0)}

	isGlobalOp := false
	for userID, user := range users {
		if currentUserID == userID {
			isGlobalOp = user.IsGlobalOps()
			break
		}
	}

	userIds := make([]string, 0)
	for userID, user := range users {
		if !isGlobalOp && user.ID != currentUserID {
			continue
		}
		userIds = append(userIds, userID)
	}

	sort.Strings(userIds)

	for _, userID := range userIds {
		r := make([]rEndpoint, 0)

		endpoints, err := deployment.Endpoints(state.Policy, state.State, userID)
		if err != nil {
			panic(err)
		}

		for key, links := range endpoints {
			instance := state.State.ResolvedData.ComponentInstanceMap[key]
			rLinks := make([]rLink, 0)

			for linkName, link := range links {
				rLinks = append(rLinks, rLink{linkName, link})
			}

			r = append(r, rEndpoint{instance.Key.ServiceName, instance.Key.ContextName, instance.Key.ContextNameWithKeys, instance.Key.ComponentName, rLinks})
		}

		uR.Endpoints = append(uR.Endpoints, userEndpoints{users[userID], r})
	}

	return uR
}
