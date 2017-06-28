package language

import (
	. "github.com/Frostman/aptomi/pkg/slinga/language/yaml"
)

// Loads service from file
func loadServiceFromFile(fileName string) *Service {
	return LoadObjectFromFile(fileName, new(Service)).(*Service)
}

// Loads context from file
func loadContextFromFile(fileName string) *Context {
	return LoadObjectFromFile(fileName, new(Context)).(*Context)
}

// Loads cluster from file
func loadClusterFromFile(fileName string) *Cluster {
	return LoadObjectFromFile(fileName, new(Cluster)).(*Cluster)
}

// Loads rules from file
func loadRulesFromFile(fileName string) []*Rule {
	return *LoadObjectFromFileDefaultEmpty(fileName, &[]*Rule{}).(*[]*Rule)
}

// Loads secrets from file
func loadUserSecretsFromFile(fileName string) []*UserSecrets {
	return *LoadObjectFromFileDefaultEmpty(fileName, &[]*UserSecrets{}).(*[]*UserSecrets)
}

// Loads dependencies from file
func loadDependenciesFromFile(fileName string) []*Dependency {
	return *LoadObjectFromFileDefaultEmpty(fileName, &[]*Dependency{}).(*[]*Dependency)
}

// Loads users from file
func loadUsersFromFile(fileName string) []*User {
	return *LoadObjectFromFileDefaultEmpty(fileName, &[]*User{}).(*[]*User)
}
