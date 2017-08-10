package language

// Context for a given service
type Context struct {
	*SlingaObject

	Criteria   *Criteria
	Labels     *LabelOperations
	Allocation *Allocation
}

// Matches checks if context criteria is satisfied
func (context *Context) Matches(labels LabelSet) bool {
	return context.Criteria == nil || context.Criteria.allows(labels)
}

func (context *Context) GetObjectType() SlingaObjectType {
	return TypePolicy
}
