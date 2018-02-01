package secrets

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"github.com/patrickmn/go-cache"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// SecretLoaderFromDir allows to load secrets for users from a given directory
type SecretLoaderFromDir struct {
	baseDir string
	cache   *cache.Cache
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
		cache:   cache.New(time.Minute, time.Minute),
	}
}

// LoadSecretsAll loads all secrets
func (loader *SecretLoaderFromDir) LoadSecretsAll() map[string]map[string]string {
	// this can be called concurrently by the engine, so it needs to be thread safe
	cachedSecrets, _ := loader.cache.Get("secrets")
	if cachedSecrets != nil {
		return cachedSecrets.(map[string]map[string]string)
	}

	// synchronize and retrieve secrets
	mutex := sync.Mutex{}
	mutex.Lock()
	defer func() { mutex.Unlock() }()

	// retrieve secrets
	result := make(map[string]map[string]string)

	if len(loader.baseDir) <= 0 {
		// log.Warnf("Skip loading secrets because baseDir not specified")
		return result
	}

	pattern := filepath.Join(loader.baseDir, "**", "secrets*.yaml")
	files, err := zglob.Glob(pattern)
	if err != nil {
		panic(fmt.Errorf("error while searching secrets files"))
	}

	sort.Strings(files)
	for _, f := range files {
		secrets := loadUserSecretsFromFile(f)
		for _, secret := range secrets {
			result[strings.ToLower(secret.User)] = secret.Secrets
		}
	}

	loader.cache.Set("secrets", result, cache.DefaultExpiration)
	return result
}

// LoadSecretsByUserName loads secrets for a single user
func (loader *SecretLoaderFromDir) LoadSecretsByUserName(user string) map[string]string {
	return loader.LoadSecretsAll()[strings.ToLower(user)]
}

// Loads secrets from file
func loadUserSecretsFromFile(fileName string) []*UserSecrets {
	log.Debugf("Loading secrets from file: %s", fileName)
	return *yaml.LoadObjectFromFileDefaultEmpty(fileName, &[]*UserSecrets{}).(*[]*UserSecrets)
}
