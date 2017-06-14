package slinga

import (
	log "github.com/Sirupsen/logrus"
	"os"
)

type AptomiOject string

const (
	// These objects can be added to Aptomi
	Clusters     AptomiOject = "policy/clusters"
	Services     AptomiOject = "policy/services"
	Contexts     AptomiOject = "policy/contexts"
	Rules        AptomiOject = "policy/rules"
	Dependencies AptomiOject = "dependencies"

	// These objects must be configured to point to external resources
	Users   AptomiOject = "external/users"
	Secrets AptomiOject = "external/secrets"
	Charts  AptomiOject = "external/charts"

	// These are generated resolution data
	PolicyResolution AptomiOject = "resolution/usage"
	Logs             AptomiOject = "resolution/logs"
	Graphics         AptomiOject = "resolution/graphics"
)

var AptomiObjectsCanBeAdded = map[string]AptomiOject{
	"cluster":      Clusters,
	"service":      Services,
	"context":      Contexts,
	"rules":        Rules,
	"dependencies": Dependencies,
	"users":        Users,
	"secrets":      Secrets,
	"chart" :       Charts,
}

// Return aptomi DB directory
func getAptomiEnvVarAsDir(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		debug.WithFields(log.Fields{
			"var": key,
		}).Fatal("Environment variable is not present. Must point to a directory")
	}
	if stat, err := os.Stat(value); err != nil || !stat.IsDir() {
		debug.WithFields(log.Fields{
			"var":       key,
			"directory": value,
			"error":     err,
		}).Fatal("Directory doesn't exist or error encountered")
	}
	return value
}

func GetAptomiBaseDir() string {
	return getAptomiEnvVarAsDir("APTOMI_DB")
}

func GetAptomiObjectDir(baseDir string, apt AptomiOject) string {
	dir := baseDir + "/" + string(apt)
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		_ = os.MkdirAll(dir, 0755)
	}
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		debug.WithFields(log.Fields{
			"directory": dir,
			"error":     err,
		}).Fatal("Directory can't be created or error encountered")
	}
	return dir
}
