package visibility

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine"
	"sort"
)

type item struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}

// TODO: change UserId -> UserId (and don't break UI...)
type detail struct {
	UserId          string
	Users           []*item
	Services        []*item
	Dependencies    []*item
	AllDependencies []*item
	Views           []*item
	Summary         engine.ServiceUsageStateSummary
}

// NewDetails returns detail object
func NewDetails(userID string, state engine.ServiceUsageState) detail {
	summary := state.GetSummary()
	r := detail{
		userID,
		make([]*item, 0),
		make([]*item, 0),
		make([]*item, 0),
		make([]*item, 0),
		make([]*item, 0),
		summary,
	}

	// Users
	userIds := make([]string, 0)
	for userID := range state.GetUserLoader().LoadUsersAll().Users {
		userIds = append(userIds, userID)
	}

	sort.Strings(userIds)

	if len(userIds) > 1 {
		r.Users = append([]*item{{"all", "All"}}, r.Users...)
	}
	for _, userID := range userIds {
		r.Users = append(r.Users, &item{userID, state.GetUserLoader().LoadUserByID(userID).Name})
	}

	// Dependencies
	depIds := make([]string, 0)
	deps := state.Dependencies.DependenciesByID
	for depID, dep := range deps {
		if dep.UserID != userID {
			continue
		}

		depIds = append(depIds, depID)
	}

	sort.Strings(depIds)

	if len(depIds) > 1 {
		r.Dependencies = append([]*item{{"all", "All"}}, r.Dependencies...)
	}
	for _, depID := range depIds {
		r.Dependencies = append(r.Dependencies, &item{depID, deps[depID].ID})
	}

	allDepIds := make([]string, 0)
	for depID := range deps {
		allDepIds = append(allDepIds, depID)
	}

	sort.Strings(allDepIds)

	if len(allDepIds) > 1 {
		r.AllDependencies = append([]*item{{"all", "All"}}, r.AllDependencies...)
	}
	for _, depID := range allDepIds {
		r.AllDependencies = append(r.AllDependencies, &item{depID, deps[depID].ID})
	}

	// Services
	svcIds := make([]string, 0)
	for svcID, svc := range state.Policy.Services {
		if svc.Owner != userID {
			continue
		}
		svcIds = append(svcIds, svcID)
	}

	sort.Strings(svcIds)

	for _, svcID := range svcIds {
		r.Services = append(r.Services, &item{svcID, state.Policy.Services[svcID].Name})
	}

	if len(r.Dependencies) > 0 {
		r.Views = append(r.Views, &item{"consumer", "Service Consumer"})
	}
	if len(r.Services) > 0 {
		r.Views = append(r.Views, &item{"service", "Service Owner"})
	}
	if state.GetUserLoader().LoadUserByID(userID).Labels["global_ops"] == "true" {
		r.Views = append(r.Views, &item{"globalops", "Global IT/Ops"})
	}

	return r
}
