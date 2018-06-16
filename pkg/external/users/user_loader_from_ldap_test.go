package users

import (
	"strings"
	"testing"

	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/stretchr/testify/assert"
)

var integrationTestsLDAP = config.LDAP{
	Host:         "localhost",
	Port:         10389,
	BaseDN:       "o=aptomiOrg",
	Filter:       "(&(objectClass=organizationalPerson))",
	FilterByName: "(&(objectClass=organizationalPerson)(cn=%s))",
	LabelToAttributes: map[string]string{
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

func TestUserLoaderFromLDAP(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	userLoaderDir := NewUserLoaderFromFile("../../testdata/ldap/users.yaml", make(map[string]bool))
	userLoaderLDAP := NewUserLoaderFromLDAP(integrationTestsLDAP, make(map[string]bool))

	usersDir := userLoaderDir.LoadUsersAll()
	usersLDAP := userLoaderLDAP.LoadUsersAll()
	assert.Equal(t, len(usersDir.Users), len(usersLDAP.Users), "Correct number of users should be loaded from LDAP")

	for _, uDir := range usersDir.Users {
		// test when loading all users
		{
			uLDAP := usersLDAP.Users[strings.ToLower(uDir.Name)]
			if !assert.NotNil(t, uLDAP, "LDAP user %s should be found", uDir.Name) {
				continue
			}
			// check that user info is correctly loaded
			compareUsers(t, uDir, uLDAP)
		}

		// test loading by name
		{
			uLDAP := userLoaderLDAP.LoadUserByName(strings.ToUpper(uDir.Name))

			// check that user is found
			if !assert.NotNil(t, uLDAP, "LDAP user %s should be found", uDir.Name) {
				continue
			}
			// check that user info is correctly loaded
			compareUsers(t, uDir, uLDAP)
		}

		// successful authentication (password == lowercase name for test users, e.g. 'alice')
		{
			user, err := userLoaderLDAP.Authenticate(uDir.Name, strings.ToLower(uDir.Name))
			assert.NoError(t, err, "Authentication should be successful")
			assert.NotEmpty(t, user, "User should be returned as a result of authentication")
		}
		// successful authentication (user names are not case sensitive)
		{
			user, err := userLoaderLDAP.Authenticate(strings.ToUpper(uDir.Name), strings.ToLower(uDir.Name))
			assert.NoError(t, err, "Authentication should be successful (user name not case sensitive)")
			assert.NotEmpty(t, user, "User should be returned as a result of authentication")
		}
		// failed authentication
		{
			user, err := userLoaderLDAP.Authenticate(uDir.Name, uDir.Name+"pass")
			assert.Error(t, err, "Authentication should not be successful")
			assert.Empty(t, user, "User should not be returned as a result of failed authentication")
		}

	}

	// check that summary is correct
	assert.Equal(t, "6 (from LDAP)", userLoaderLDAP.Summary())
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
