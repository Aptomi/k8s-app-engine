package users

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"gopkg.in/ldap.v2"
	"strconv"
	"strings"
	"sync"
)

// UserLoaderFromLDAP allows aptomi to load users from LDAP
type UserLoaderFromLDAP struct {
	once sync.Once

	cfg   config.LDAP
	users *lang.GlobalUsers
}

// NewUserLoaderFromLDAP returns new UserLoaderFromLDAP, given location with LDAP configuration file (with host/port and mapping)
func NewUserLoaderFromLDAP(cfg config.LDAP) UserLoader {
	return &UserLoaderFromLDAP{
		cfg: cfg,
	}
}

// LoadUsersAll loads all users
func (loader *UserLoaderFromLDAP) LoadUsersAll() *lang.GlobalUsers {
	// Right now this can be called concurrently by the engine, so it needs to be thread safe
	loader.once.Do(func() {
		loader.users = &lang.GlobalUsers{Users: make(map[string]*lang.User)}
		t := loader.ldapSearch()
		for _, u := range t {
			loader.users.Users[u.ID] = u
		}
	})
	return loader.users
}

// LoadUserByID loads a single user by ID
func (loader *UserLoaderFromLDAP) LoadUserByID(id string) *lang.User {
	return loader.LoadUsersAll().Users[id]
}

// Summary returns summary as string
func (loader *UserLoaderFromLDAP) Summary() string {
	return strconv.Itoa(len(loader.LoadUsersAll().Users)) + " (from LDAP)"
}

// Does search on LDAP and returns entries
func (loader *UserLoaderFromLDAP) ldapSearch() []*lang.User {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", loader.cfg.Host, loader.cfg.Port))
	if err != nil {
		panic(err)
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		loader.cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		loader.cfg.Filter,
		loader.cfg.GetAttributes(),
		nil,
	)

	searchResult, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}

	result := []*lang.User{}
	for _, entry := range searchResult.Entries {
		user := &lang.User{
			ID:     entry.DN,
			Name:   entry.GetAttributeValue(loader.cfg.LabelToAttributes["name"]),
			Labels: make(map[string]string),
		}
		for label, attr := range loader.cfg.LabelToAttributes {
			if label != "id" && label != "name" {
				value := entry.GetAttributeValue(attr)
				if len(value) > 0 {
					user.Labels[label] = ldapValue(value)
				}
			}
		}

		// fmt.Printf("%+v\n", user)
		result = append(result, user)
	}

	return result
}

func ldapValue(value string) string {
	// normalize boolean values
	if strings.ToLower(value) == "true" {
		return "true"
	}
	if strings.ToLower(value) == "false" {
		return "false"
	}
	return value
}
