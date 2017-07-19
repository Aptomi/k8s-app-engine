package visibility

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"sort"
)

type rLink struct {
	Name string
	Link string
}

type rEndpoint struct {
	Service    string
	Context    string
	Allocation string
	Component  string
	Links      []rLink
}

type userEndpoints struct {
	User      *User
	Endpoints []rEndpoint
}

type endpointsView struct {
	Endpoints []userEndpoints
}

// Endpoints returns a view with all endpoints
func Endpoints(currentUserID string, state engine.ServiceUsageState) endpointsView {
	users := state.GetUserLoader().LoadUsersAll().Users

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

		endpoints := state.Endpoints(userID)

		for key, links := range endpoints {
			service, context, allocation, component := engine.ParseServiceUsageKey(key)
			rLinks := make([]rLink, 0)

			for linkName, link := range links {
				rLinks = append(rLinks, rLink{linkName, link})
			}

			r = append(r, rEndpoint{service, context, allocation, component, rLinks})
		}

		uR.Endpoints = append(uR.Endpoints, userEndpoints{users[userID], r})
	}

	return uR
}
