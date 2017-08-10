package language

func loadUnitTestsPolicy() *Policy {
	db := NewSlingaObjectDatabaseDir("../testdata/unittests_new")
	return db.LoadPolicyObjects(-1, "")
}
