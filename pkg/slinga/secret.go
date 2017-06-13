package slinga

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

// LoadUserSecretsByIDFromDir loads secrets for a given user from a given directory
func LoadUserSecretsByIDFromDir(dir string, id string) *LabelSet {
	fileName := dir + "/secrets.yaml"
	debug.WithFields(log.Fields{
		"file": fileName,
	}).Debug("Loading secrets")

	dat, e := ioutil.ReadFile(fileName)

	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}
	t := []*UserSecrets{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal secrets")
	}
	for _, s := range t {
		if s.UserID == id {
			return &LabelSet{Labels: s.Secrets}
		}
	}

	return &LabelSet{Labels: make(map[string]string)}
}
