package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
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
	User      *slinga.User
	Endpoints []rEndpoint
}

type endpointsView struct {
	Endpoints []userEndpoints
}

func Endpoints(userID string, users slinga.GlobalUsers, state slinga.ServiceUsageState) endpointsView {
	uR := endpointsView{make([]userEndpoints, 0)}

	userIds := make([]string, 0)
	for userID, user := range users.Users {
		if !user.IsGlobalOps() && user.ID != userID {
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
