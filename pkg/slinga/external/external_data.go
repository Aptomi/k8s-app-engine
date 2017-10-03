package external

import (
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
)

// Data represents all data which is external to Aptomi, including
// - users
// - secrets
type Data struct {
	UserLoader   users.UserLoader
	SecretLoader secrets.SecretLoader
}

// NewData creates a new instance of external Data
func NewData(userLoader users.UserLoader, secretLoader secrets.SecretLoader) *Data {
	return &Data{
		UserLoader:   userLoader,
		SecretLoader: secretLoader,
	}
}
