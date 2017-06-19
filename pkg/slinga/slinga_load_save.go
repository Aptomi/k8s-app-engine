package slinga

// Loads service from file
func loadServiceFromFile(fileName string) *Service {
	return loadObjectFromFile(fileName, new(Service)).(*Service)
}

// Loads context from file
func loadContextFromFile(fileName string) *Context {
	return loadObjectFromFile(fileName, new(Context)).(*Context)
}

// Loads cluster from file
func loadClusterFromFile(fileName string) *Cluster {
	return loadObjectFromFile(fileName, new(Cluster)).(*Cluster)
}

// Loads revision from file
func loadRevisionFromFile(fileName string) AptomiRevision {
	return *loadObjectFromFile(fileName, new(AptomiRevision)).(*AptomiRevision)
}

// Loads usage state from file
func loadServiceUsageStateFromFile(fileName string) ServiceUsageState {
	return *loadObjectFromFileDefaultEmpty(fileName, new(ServiceUsageState)).(*ServiceUsageState)
}

// Loads rules from file
func loadRulesFromFile(fileName string) []*Rule {
	return *loadObjectFromFileDefaultEmpty(fileName, &[]*Rule{}).(*[]*Rule)
}

// Loads secrets from file
func loadUserSecretsFromFile(fileName string) []*UserSecrets {
	return *loadObjectFromFileDefaultEmpty(fileName, &[]*UserSecrets{}).(*[]*UserSecrets)
}

// Loads dependencies from file
func loadDependenciesFromFile(fileName string) []*Dependency {
	return *loadObjectFromFileDefaultEmpty(fileName, &[]*Dependency{}).(*[]*Dependency)
}

// Loads users from file
func loadUsersFromFile(fileName string) []*User {
	return *loadObjectFromFileDefaultEmpty(fileName, &[]*User{}).(*[]*User)
}
