package slinga

import (
	"github.com/golang/glog"
	"os"
)

// Return aptomi DB directory
func getAptomiEnvVarAsDir(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		glog.Fatalf("%s environment variable is not present. Must point to a directory", key)
	}
	if stat, err := os.Stat(value); err != nil || !stat.IsDir() {
		glog.Fatalf("Directory %s doesn't exist: %s", key, value)
	}
	return value
}

// Return aptomi DB directory
func GetAptomiDBDir() string {
	return getAptomiEnvVarAsDir("APTOMI_DB")
}

// Return aptomi policy directory
func GetAptomiPolicyDir() string {
	return getAptomiEnvVarAsDir("APTOMI_POLICY")
}
