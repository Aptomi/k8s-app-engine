package secrets

// SecretLoader is an interface which allows aptomi to load secrets for users
// from different sources (e.g. file, external store, etc)
type SecretLoader interface {
	// LoadSecretsByUserID should load a set of secrets for a given user
	LoadSecretsByUserID(string) map[string]string
}
