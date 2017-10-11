package secrets

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/yaml"
	"github.com/mattn/go-zglob"
	"path/filepath"
	"sort"
	"sync"
)

// SecretLoaderFromDir allows to load secrets for users from a given directory
type SecretLoaderFromDir struct {
	once sync.Once

	baseDir       string
	cachedSecrets map[string]map[string]string
}

// UserSecrets represents a user secret (ID, set of secrets)
type UserSecrets struct {
	UserID  string
	Secrets map[string]string
}

// NewSecretLoaderFromDir returns new UserLoaderFromDir, given a directory where files should be read from
func NewSecretLoaderFromDir(baseDir string) SecretLoader {
	return &SecretLoaderFromDir{
		baseDir: baseDir,
	}
}

// LoadSecretsAll loads all secrets
func (loader *SecretLoaderFromDir) LoadSecretsAll() map[string]map[string]string {
	// Right now this can be called concurrently by the engine, so it needs to be thread safe
	loader.once.Do(func() {
		loader.cachedSecrets = make(map[string]map[string]string)

		pattern := filepath.Join(loader.baseDir, "**", "secrets*.yaml")
		files, err := zglob.Glob(pattern)
		if err != nil {
			panic(fmt.Errorf("error while searching secrets files"))
		}
		sort.Strings(files)
		for _, f := range files {
			secrets := loadUserSecretsFromFile(f)
			for _, secret := range secrets {
				loader.cachedSecrets[secret.UserID] = secret.Secrets
			}
		}
	})
	return loader.cachedSecrets
}

// LoadSecretsByUserID loads secrets for a single user
func (loader *SecretLoaderFromDir) LoadSecretsByUserID(userID string) map[string]string {
	return loader.LoadSecretsAll()[userID]
}

// Loads secrets from file
func loadUserSecretsFromFile(fileName string) []*UserSecrets {
	return *yaml.LoadObjectFromFileDefaultEmpty(fileName, &[]*UserSecrets{}).(*[]*UserSecrets)
}
