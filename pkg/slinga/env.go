package slinga

import (
	log "github.com/Sirupsen/logrus"
	"os"
)

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
		}).Fatal("Directory doesn't exist")
	}
	return value
}

// GetAptomiDBDir returns Aptomi DB directory
func GetAptomiDBDir() string {
	return getAptomiEnvVarAsDir("APTOMI_DB")
}

// GetAptomiPolicyDir returns Aptomi Policy directory
func GetAptomiPolicyDir() string {
	return getAptomiEnvVarAsDir("APTOMI_POLICY")
}
