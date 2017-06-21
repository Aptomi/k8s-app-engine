package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
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

type endpointsView struct {
	Endpoints map[string][]rEndpoint
}

func Endpoints(filterUserId string, users map[string]*slinga.User, state slinga.ServiceUsageState) endpointsView {
	uR := endpointsView{make(map[string][]rEndpoint)}

	for userId := range users {
		if filterUserId != "" && userId != filterUserId {
			continue
		}
		r := make([]rEndpoint, 0)

		endpoints := state.Endpoints(userId)

		for key, links := range endpoints {
			service, context, allocation, component := slinga.ParseServiceUsageKey(key)
			rLinks := make([]rLink, 0)

			for linkName, link := range links {
				rLinks = append(rLinks, rLink{linkName, link})
			}

			r = append(r, rEndpoint{service, context, allocation, component, rLinks})
		}

		uR.Endpoints[userId] = r
	}

	return uR
}
