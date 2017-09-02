package external

import (
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
)

// Data represents all data which is external to Aptomi, including
// - users
type Data struct {
	UserLoader users.UserLoader
}

func NewData(userLoader users.UserLoader) *Data {
	return &Data{
		UserLoader: userLoader,
	}
}
