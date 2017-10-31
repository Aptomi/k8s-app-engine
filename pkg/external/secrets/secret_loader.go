package secrets

// SecretLoader is an interface which allows aptomi to load secrets for users
// from different sources (e.g. file, external store, etc)
type SecretLoader interface {
	// LoadSecretsByUserName should load a set of secrets for a given user
	LoadSecretsByUserName(string) map[string]string
}
