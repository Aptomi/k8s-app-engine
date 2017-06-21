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
	Users        []*item
	Services     []*item
	Dependencies []*item
}

func NewDetails(userId string, globalUsers slinga.GlobalUsers, state slinga.ServiceUsageState) detail {
	r := detail{make([]*item, 0), make([]*item, 0), make([]*item, 0)}

	// Users
	userIds := make([]string, 0)
	for userId := range globalUsers.Users {
		userIds = append(userIds, userId)
	}

	sort.Strings(userIds)

	for _, userId := range userIds {
		r.Users = append(r.Users, &item{userId, globalUsers.Users[userId].Name})
	}

	return r
}
