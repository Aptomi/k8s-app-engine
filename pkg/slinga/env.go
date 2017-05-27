package slinga

import (
	"os"
)

// Return aptomi DB directory
func getAptomiEnvVarAsDir(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		debug.Fatalf("%s environment variable is not present. Must point to a directory", key)
	}
	if stat, err := os.Stat(value); err != nil || !stat.IsDir() {
		debug.Fatalf("Directory %s doesn't exist: %s", key, value)
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
