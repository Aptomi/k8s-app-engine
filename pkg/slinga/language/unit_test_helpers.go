package language

func loadUnitTestsPolicy() *PolicyNamespace {
	db := NewSlingaObjectDatabaseDir("../testdata/unittests")
	return db.LoadPolicyObjects(-1, "")
}
