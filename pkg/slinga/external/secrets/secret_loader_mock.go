package secrets

// SecretLoaderMock allows to mock secret loader and use in-memory user storage
type SecretLoaderMock struct {
	secrets map[string]map[string]string
}

// NewSecretLoaderMock returns new SecretLoaderMock
func NewSecretLoaderMock() *SecretLoaderMock {
	return &SecretLoaderMock{
		secrets: make(map[string]map[string]string),
	}
}

// AddSecret adds a secret for a given user
func (loader *SecretLoaderMock) AddSecret(userID string, name string, value string) {
	loader.secrets[userID][name] = value
}

// LoadSecretsAll loads all secrets
func (loader *SecretLoaderMock) LoadSecretsAll() map[string]map[string]string {
	return loader.secrets
}

// LoadSecretsByUserID loads secrets for a single user
func (loader *SecretLoaderMock) LoadSecretsByUserID(userID string) map[string]string {
	return loader.secrets[userID]
}
