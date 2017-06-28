package language

import (
	. "github.com/Frostman/aptomi/pkg/slinga/db"
	"github.com/Frostman/aptomi/pkg/slinga/language/yaml"
	"github.com/mattn/go-zglob"
	"sort"
)

/*
	This file declares all necessary structures for Secrets to be retrieved
	For now it loads secrets from YAML file
	Later this will be replaced with some external Secrets Storage like Vault
*/

// UserSecrets represents a user secret (ID, set of secrets)
type UserSecrets struct {
	UserID  string
	Secrets map[string]string
}

func loadUserSecretsFromDir(baseDir string) []*UserSecrets {
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeSecrets))
	sort.Strings(files)
	t := []*UserSecrets{}
	for _, f := range files {
		t = append(t, loadUserSecretsFromFile(f)...)
	}
	return t
}

// LoadUserSecretsByIDFromDir loads all secrets for a particular user
func LoadUserSecretsByIDFromDir(baseDir string, id string) map[string]string {
	t := loadUserSecretsFromDir(baseDir)
	for _, s := range t {
		if s.UserID == id {
			return s.Secrets
		}
	}

	return make(map[string]string)
}

// Loads secrets from file
func loadUserSecretsFromFile(fileName string) []*UserSecrets {
	return *yaml.LoadObjectFromFileDefaultEmpty(fileName, &[]*UserSecrets{}).(*[]*UserSecrets)
}
