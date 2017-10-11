package users

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/config"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/stretchr/testify/assert"
	"testing"
)

var integrationTestsLDAP = config.LDAP{
	Host:   "localhost",
	Port:   10389,
	BaseDN: "o=aptomiOrg",
	Filter: "(&(objectClass=organizationalPerson))",
	LabelToAttributes: map[string]string{
		"id":                "dn",
		"name":              "cn",
		"description":       "description",
		"global_ops":        "isglobalops",
		"is_operator":       "isoperator",
		"mail":              "mail",
		"team":              "team",
		"org":               "o",
		"short-description": "role",
		"deactivated":       "deactivated",
	},
}

func TestLoadUsersFromLDAP(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	userLoaderDir := NewUserLoaderFromDir("../../testdata/integrationtests")
	userLoaderLDAP := NewUserLoaderFromLDAP(integrationTestsLDAP)

	usersDir := userLoaderDir.LoadUsersAll()
	usersLDAP := userLoaderLDAP.LoadUsersAll()
	assert.Equal(t, len(usersDir.Users), len(usersLDAP.Users), "Correct number of users should be loaded from LDAP")

	for _, uDir := range usersDir.Users {
		id := uDir.Labels["ldapDN"]
		uLDAP := usersLDAP.Users[id]

		// check that user is found
		if !assert.NotNil(t, uLDAP, fmt.Sprintf("LDAP user %s should be found", id)) {
			continue
		}

		// check that user info is correctly loaded
		compareUsers(t, uDir, uLDAP)
	}

	// check that summary is correct
	assert.Equal(t, "6 (from LDAP)", userLoaderLDAP.Summary())
}

func TestLoadUsersFromLDAPIndividually(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	userLoaderDir := NewUserLoaderFromDir("../../testdata/integrationtests")
	userLoaderLDAP := NewUserLoaderFromLDAP(integrationTestsLDAP)

	usersDir := userLoaderDir.LoadUsersAll()

	for _, uDir := range usersDir.Users {
		id := uDir.Labels["ldapDN"]
		uLDAP := userLoaderLDAP.LoadUserByID(id)

		// check that user is found
		if !assert.NotNil(t, uLDAP, fmt.Sprintf("LDAP user %s should be found", id)) {
			continue
		}

		// check that user info is correctly loaded
		compareUsers(t, uDir, uLDAP)
	}
}

func compareUsers(t *testing.T, uDir *lang.User, uLDAP *lang.User) {
	// check that name matches
	assert.Equal(t, uDir.Name, uLDAP.Name, "User LDAP name should match")
	// check that labels are mapped correctly
	for key, valueDir := range uDir.Labels {
		if key != "ldapDN" {
			valueLDAP := uLDAP.Labels[key]
			assert.Equal(t, valueDir, valueLDAP, "User label mapped from LDAP should match")
		}
	}
}
