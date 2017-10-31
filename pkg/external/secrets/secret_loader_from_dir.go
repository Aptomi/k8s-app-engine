package secrets

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
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

// UserSecrets represents a single user secret (user name and a map of secrets)
type UserSecrets struct {
	User    string
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
				loader.cachedSecrets[secret.User] = secret.Secrets
			}
		}
	})
	return loader.cachedSecrets
}

// LoadSecretsByUserName loads secrets for a single user
func (loader *SecretLoaderFromDir) LoadSecretsByUserName(userName string) map[string]string {
	return loader.LoadSecretsAll()[userName]
}

// Loads secrets from file
func loadUserSecretsFromFile(fileName string) []*UserSecrets {
	return *yaml.LoadObjectFromFileDefaultEmpty(fileName, &[]*UserSecrets{}).(*[]*UserSecrets)
}
