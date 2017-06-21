package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
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
	for userId, user := range slinga.LoadUsers().Users {
		r.Users = append(r.Users, &item{userId, user.Name})
	}

	return r
}
