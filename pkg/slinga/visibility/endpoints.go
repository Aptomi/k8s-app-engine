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

func Endpoints(userID string, users map[string]*slinga.User, state slinga.ServiceUsageState) endpointsView {
	uR := endpointsView{make([]userEndpoints, 0)}

	isGlobalOp := users[userID].Labels["global_ops"] == "true"

	userIds := make([]string, 0)
	for userID, user := range users {
		if !isGlobalOp && user.ID != userID {
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

		uR.Endpoints = append(uR.Endpoints, userEndpoints{users[userID], r})
	}

	return uR
}
