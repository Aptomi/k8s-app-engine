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
	Views        []*item
}

func NewDetails(userId string, globalUsers slinga.GlobalUsers, state slinga.ServiceUsageState) detail {
	r := detail{userId, make([]*item, 0), make([]*item, 0), make([]*item, 0), make([]*item, 0)}

	// Users
	userIds := make([]string, 0)
	for userId := range globalUsers.Users {
		userIds = append(userIds, userId)
	}

	sort.Strings(userIds)

	if len(userIds) > 1 {
		r.Users = append([]*item{{"all", "All"}}, r.Users...)
	}
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

	if len(depIds) > 1 {
		r.Dependencies = append([]*item{{"all", "All"}}, r.Dependencies...)
	}
	for _, depId := range depIds {
		r.Dependencies = append(r.Dependencies, &item{depId, deps[depId].ID})
	}

	// Services
	svcIds := make([]string, 0)
	for svcId, svc := range state.Policy.Services {
		if svc.Owner != userId {
			continue
		}
		svcIds = append(svcIds, svcId)
	}

	sort.Strings(svcIds)

	for _, svcId := range svcIds {
		r.Services = append(r.Services, &item{svcId, state.Policy.Services[svcId].Name})
	}

	if len(r.Dependencies) > 0 {
		r.Views = append(r.Views, &item{"consumer", "Service Consumer View"})
	}
	if len(r.Services) > 0 {
		r.Views = append(r.Views, &item{"service", "Service Owner View"})
	}
	if globalUsers.Users[userId].Labels["global_ops"] == "true" {
		r.Views = append(r.Views, &item{"globalops", "Global IT/Ops View"})
	}

	return r
}
