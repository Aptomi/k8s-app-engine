package language

// SlingaObjectDatabase is an interface that allows CRUD operations on aptomi objects
type SlingaObjectDatabase interface {
	LoadPolicyObjects(revision int, namespace string) *Policy
}

// TODO: we have policy objects and calculated objects. API must support loading all kinds
