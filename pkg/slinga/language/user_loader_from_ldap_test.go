package language

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadUsersFromLDAP(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	userLoaderDir := NewUserLoaderFromDir("../testdata/integrationtests")
	userLoaderLDAP := NewUserLoaderFromLDAP("../testdata/integrationtests")

	usersDir := userLoaderDir.LoadUsersAll()
	usersLDAP := userLoaderLDAP.LoadUsersAll()
	assert.Equal(t, len(usersDir.Users), len(usersLDAP.Users), "Correct number of users should be loaded from LDAP")

	for _, uDir := range usersDir.Users {
		id := uDir.Labels["ldapDN"]
		uLDAP := usersLDAP.Users[id]
		assert.NotNil(t, uLDAP, fmt.Sprintf("LDAP user %s should be found", id))

		assert.Equal(t, uDir.Name, uLDAP.Name, "User LDAP name should match")

		for key, valueDir := range uDir.Labels {
			if key != "ldapDN" {
				valueLDAP := uLDAP.Labels[key]
				assert.Equal(t, valueDir, valueLDAP, "User label mapped from LDAP should match")
			}
		}
	}
}
