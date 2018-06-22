package dbtest_test

import (
	"testing"

	"github.com/Aptomi/aptomi/pkg/runtime/db"
	_ "github.com/Aptomi/aptomi/pkg/runtime/db/driver/bolt"
	"github.com/stretchr/testify/assert"
)

func TestBasicDatabaseUsage(t *testing.T) {
	// todo should we have "test suite" in db package and import it from each db to test it? what's about benchmark?
	// todo convert to table test + forEachDb ...
	t.Run("Bolt", func(tt *testing.T) {
		store, err := db.Open("bolt", "tbd")
		assert.NoError(tt, err, "should be able to open DB")
		assert.NotNil(tt, store, "store should be not nil")
	})
}
