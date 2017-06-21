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
	Endpoints []rEndpoint
}

func Endpoints(filterUserId string, users map[string]*slinga.User, state slinga.ServiceUsageState) endpointsView {
	r := make([]rEndpoint, 0)

	for userId := range users {
		if filterUserId != "" && userId != filterUserId {
			continue
		}

		endpoints := state.Endpoints(userId)

		for key, links := range endpoints {
			service, context, allocation, component := slinga.ParseServiceUsageKey(key)
			rLinks := make([]rLink, 0)

			for linkName, link := range links {
				rLinks = append(rLinks, rLink{linkName, link})
			}

			r = append(r, rEndpoint{service, context, allocation, component, rLinks})
		}

	}

	return endpointsView{r}
}
