package language

func loadUnitTestsPolicy() *Policy {
	db := NewSlingaObjectDatabaseDir("../testdata/unittests")
	return db.LoadPolicyObjects(-1, "")
}
