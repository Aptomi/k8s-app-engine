package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
)

func loadUnitTestsPolicy() *Policy {
	db := NewSlingaObjectDatabaseDir("../testdata/unittests_new")
	return db.LoadPolicyObjects(-1, "")
}
