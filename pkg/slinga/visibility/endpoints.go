package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	. "github.com/Frostman/aptomi/pkg/slinga/language"
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
func Endpoints(currentUserID string, users GlobalUsers, state slinga.ServiceUsageState) endpointsView {
	uR := endpointsView{make([]userEndpoints, 0)}

	isGlobalOp := false
	for userID, user := range users.Users {
		if currentUserID == userID {
			isGlobalOp = user.IsGlobalOps()
			break
		}
	}

	userIds := make([]string, 0)
	for userID, user := range users.Users {
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
			service, context, allocation, component := slinga.ParseServiceUsageKey(key)
			rLinks := make([]rLink, 0)

			for linkName, link := range links {
				rLinks = append(rLinks, rLink{linkName, link})
			}

			r = append(r, rEndpoint{service, context, allocation, component, rLinks})
		}

		uR.Endpoints = append(uR.Endpoints, userEndpoints{users.Users[userID], r})
	}

	return uR
}
