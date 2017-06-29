package language

import (
	"strconv"
	"gopkg.in/ldap.v2"
	"fmt"
	"github.com/Frostman/aptomi/pkg/slinga/language/yaml"
	. "github.com/Frostman/aptomi/pkg/slinga/db"
	. "github.com/Frostman/aptomi/pkg/slinga/util"
	. "github.com/Frostman/aptomi/pkg/slinga/log"
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"strings"
)

type LDAPConfig struct {
	Host              string
	Port              int
	BaseDN            string
	Filter            string
	LabelToAtrributes map[string]string
}

// Loads LDAP configuration
func loadLDAPConfig(baseDir string) *LDAPConfig {
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeUsersLDAP))
	fileName, err := EnsureSingleFile(files)
	if err != nil {
		Debug.WithFields(log.Fields{
			"error": err,
		}).Panic("LDAP config lookup error")
	}
	result := yaml.LoadObjectFromFile(fileName, &LDAPConfig{}).(*LDAPConfig)

	Debug.WithFields(log.Fields{
		"config": fmt.Sprintf("%v", result),
	}).Info("Loaded LDAP config and mappings")

	return result
}

// Returns the list of attributes to be retrieved from LDAP
func (config *LDAPConfig) getAttributes() []string {
	result := []string{}
	for _, attr := range config.LabelToAtrributes {
		result = append(result, attr)
	}
	return result
}

// UserLoaderFromLDAP allows aptomi to load users from LDAP
type UserLoaderFromLDAP struct {
	baseDir     string
	config      *LDAPConfig
	cachedUsers *GlobalUsers
}

// NewUserLoaderFromLDAP returns new UserLoaderFromLDAP, given location with LDAP configuration file (with host/port and mapping)
func NewUserLoaderFromLDAP(baseDir string) UserLoader {
	return &UserLoaderFromLDAP{baseDir: baseDir, config: loadLDAPConfig(baseDir)}
}

// LoadUsersAll loads all users
func (loader *UserLoaderFromLDAP) LoadUsersAll() GlobalUsers {
	if loader.cachedUsers == nil {
		loader.cachedUsers = &GlobalUsers{Users: make(map[string]*User)}
		t := loader.ldapSearch()
		for _, u := range t {
			// load secrets
			u.Secrets = LoadUserSecretsByIDFromDir(loader.baseDir, u.ID)

			// add user
			loader.cachedUsers.Users[u.ID] = u
		}

	}
	return *loader.cachedUsers
}

// LoadUserByID loads a single user by ID
func (loader *UserLoaderFromLDAP) LoadUserByID(id string) *User {
	return loader.LoadUsersAll().Users[id]
}

// Summary returns summary as string
func (loader *UserLoaderFromLDAP) Summary() string {
	return strconv.Itoa(len(loader.LoadUsersAll().Users)) + " (from LDAP)"
}

// Does search on LDAP and returns entries
func (loader *UserLoaderFromLDAP) ldapSearch() []*User {
	Debug.WithFields(log.Fields{
		"host": loader.config.Host,
		"post": loader.config.Port,
	}).Info("Opening connection to LDAP")

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", loader.config.Host, loader.config.Port))
	if err != nil {
		panic(err)
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		loader.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		loader.config.Filter,
		loader.config.getAttributes(),
		nil,
	)

	Debug.WithFields(log.Fields{
		"baseDN":     loader.config.BaseDN,
		"filter":     loader.config.Filter,
		"attributes": loader.config.getAttributes(),
	}).Info("Making search request to LDAP")

	searchResult, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}

	Debug.WithFields(log.Fields{
		"count": len(searchResult.Entries),
	}).Info("Found entries")

	result := []*User{}
	for _, entry := range searchResult.Entries {
		user := &User{
			ID:     entry.DN,
			Name:   entry.GetAttributeValue(loader.config.LabelToAtrributes["name"]),
			Labels: make(map[string]string),
		}
		for label, attr := range loader.config.LabelToAtrributes {
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
