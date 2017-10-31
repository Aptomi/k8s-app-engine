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
func (loader *SecretLoaderMock) AddSecret(userName string, secretName string, secretValue string) {
	loader.secrets[userName][secretName] = secretValue
}

// LoadSecretsAll loads all secrets
func (loader *SecretLoaderMock) LoadSecretsAll() map[string]map[string]string {
	return loader.secrets
}

// LoadSecretsByUserName loads secrets for a single user
func (loader *SecretLoaderMock) LoadSecretsByUserName(userName string) map[string]string {
	return loader.secrets[userName]
}
