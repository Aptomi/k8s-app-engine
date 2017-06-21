package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	"sort"
)

type item struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}

type detail struct {
	UserId       string
	Users        []*item
	Services     []*item
	Dependencies []*item
}

func NewDetails(userId string, globalUsers slinga.GlobalUsers, state slinga.ServiceUsageState) detail {
	r := detail{userId, make([]*item, 0), make([]*item, 0), make([]*item, 0)}

	// Users
	userIds := make([]string, 0)
	for userId := range globalUsers.Users {
		userIds = append(userIds, userId)
	}

	sort.Strings(userIds)

	for _, userId := range userIds {
		r.Users = append(r.Users, &item{userId, globalUsers.Users[userId].Name})
	}

	// Dependencies
	depIds := make([]string, 0)
	deps := state.Dependencies.DependenciesByID
	for depId, dep := range deps {
		if dep.UserID != userId {
			continue
		}

		depIds = append(depIds, depId)
	}

	sort.Strings(depIds)

	r.Dependencies = append(r.Dependencies, &item{"all", "All"})
	for _, depId := range depIds {
		r.Dependencies = append(r.Dependencies, &item{depId, deps[depId].ID})
	}

	// Services
	svcIds := make([]string, 0)
	for svcId := range state.Policy.Services {
		svcIds = append(svcIds, svcId)
	}

	sort.Strings(svcIds)

	for _, svcId := range svcIds {
		// check service owner
		r.Services = append(r.Services, &item{svcId, state.Policy.Services[svcId].Name})
	}

	return r
}
