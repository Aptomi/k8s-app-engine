package slinga

import (
	"os"
	"log"
)

// Return aptomi DB directory
func getAptomiEnvVarAsDir(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatal(key + " environment variable is not present. Must point to a directory")
	}
	if stat, err := os.Stat(value); err != nil || !stat.IsDir() {
		log.Fatal("Directory " + key + " doesn't exist: " + value)
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

