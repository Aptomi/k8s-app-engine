package users

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/patrickmn/go-cache"
	"gopkg.in/ldap.v2"
)

// UserLoaderFromLDAP allows aptomi to load users from LDAP
type UserLoaderFromLDAP struct {
	cfg                  config.LDAP
	cache                *cache.Cache
	domainAdminOverrides map[string]bool
}

// NewUserLoaderFromLDAP returns new UserLoaderFromLDAP, given location with LDAP configuration file (with host/port and mapping)
func NewUserLoaderFromLDAP(cfg config.LDAP, domainAdminOverrides map[string]bool) UserLoader {
	return &UserLoaderFromLDAP{
		cfg:                  cfg,
		cache:                cache.New(time.Minute, time.Minute),
		domainAdminOverrides: domainAdminOverrides,
	}
}

// LoadUsersAll loads all users
func (loader *UserLoaderFromLDAP) LoadUsersAll() *lang.GlobalUsers {
	// this can be called concurrently by the engine, so it needs to be thread safe
	cachedUsers, _ := loader.cache.Get("ldapUsers")
	if cachedUsers != nil {
		return cachedUsers.(*lang.GlobalUsers)
	}

	// synchronize and retrieve users
	mutex := sync.Mutex{}
	mutex.Lock()
	defer func() { mutex.Unlock() }()

	result := &lang.GlobalUsers{Users: make(map[string]*lang.User)}
	ldapUsers, err := loader.ldapSearch()
	if err != nil {
		// we need user data, but they cannot be loaded from LDAP. for now, let's panic
		panic(err)
	}
	for _, u := range ldapUsers {
		result.Users[strings.ToLower(u.Name)] = u
		if _, exist := loader.domainAdminOverrides[strings.ToLower(u.Name)]; exist {
			u.DomainAdmin = true
		}
	}
	loader.cache.Set("ldapUsers", result, cache.DefaultExpiration)
	return result
}

// LoadUserByName loads a single user by name
func (loader *UserLoaderFromLDAP) LoadUserByName(name string) *lang.User {
	return loader.LoadUsersAll().Users[strings.ToLower(name)]
}

// Authenticate should authenticate a user by username/password.
// If user exists and username/password is correct, it should be returned.
// If a user doesn't exist or username/password is not correct, then nil should be returned together with an error.
func (loader *UserLoaderFromLDAP) Authenticate(name, password string) (*lang.User, error) {
	return loader.ldapAuthenticate(name, password)
}

// Summary returns summary as string
func (loader *UserLoaderFromLDAP) Summary() string {
	return strconv.Itoa(len(loader.LoadUsersAll().Users)) + " (from LDAP)"
}

// Does search on LDAP and returns entries
func (loader *UserLoaderFromLDAP) ldapSearch() ([]*lang.User, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", loader.cfg.Host, loader.cfg.Port))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	result := []*lang.User{}
	for _, entry := range searchResult.Entries {
		user := loader.userFromLDAPEntry(entry)
		result = append(result, user)
	}

	return result, nil
}

// Authenticates a user in ldap
func (loader *UserLoaderFromLDAP) ldapAuthenticate(name, password string) (*lang.User, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", loader.cfg.Host, loader.cfg.Port))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Start TLS
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}) // nolint: gas
	if err != nil {
		return nil, err
	}

	// Search for a user by name
	searchRequest := ldap.NewSearchRequest(
		loader.cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(loader.cfg.FilterByName, name),
		loader.cfg.GetAttributes(),
		nil,
	)

	searchResult, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(searchResult.Entries) <= 0 {
		return nil, fmt.Errorf("user '%s' does not exist in LDAP", name)
	}
	if len(searchResult.Entries) > 1 {
		return nil, fmt.Errorf("too many LDAP entries returned for user '%s'", name)
	}

	// Bind as the user to verify their password
	entry := searchResult.Entries[0]
	dn := entry.DN
	err = l.Bind(dn, password)
	if err != nil {
		return nil, fmt.Errorf("LDAP bind failed for user '%s': %s", name, err)
	}

	user := loader.userFromLDAPEntry(entry)
	return user, nil
}

func (loader *UserLoaderFromLDAP) userFromLDAPEntry(entry *ldap.Entry) *lang.User {
	name := entry.GetAttributeValue(loader.cfg.LabelToAttributes["name"])
	user := &lang.User{
		Name:   name,
		Labels: make(map[string]string),
	}
	for label, attr := range loader.cfg.LabelToAttributes {
		if label != "name" {
			value := entry.GetAttributeValue(attr)
			if len(value) > 0 {
				user.Labels[label] = ldapValue(value)
			}
		}
	}
	return user
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
