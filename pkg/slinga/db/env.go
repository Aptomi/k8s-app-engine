package db

import (
	"os"
	"path/filepath"
)

// AptomiObject represents an aptomi entity, which gets stored in aptomi DB
type AptomiObject string

const (
	/*
		The following objects must be configured to point to external resources
	*/

	// TypeUsersFile is where users are stored (this is for file-based storage)
	TypeUsersFile AptomiObject = "users"

	// TypeUsersLDAP is where ldap configuration is stored
	TypeUsersLDAP AptomiObject = "ldap"

	// TypeSecrets is where secret tokens are stored (later in Hashicorp Vault)
	TypeSecrets AptomiObject = "secrets"

	// TypeCharts is where binary charts/images are stored (later in external repo)
	TypeCharts AptomiObject = "charts"
)

// Return aptomi DB directory
func getAptomiEnvVarAsDir(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic("Environment variable is not present. Must point to a directory")
	}
	if stat, err := os.Stat(value); err != nil || !stat.IsDir() {
		panic("Directory doesn't exist or error encountered")
	}
	return value
}

// GetAptomiBaseDir returns base directory, i.e. the value of APTOMI_DB environment variable
func GetAptomiBaseDir() string {
	return getAptomiEnvVarAsDir("APTOMI_DB")
}

// GetAptomiPolicyDir returns default aptomi policy dir
func GetAptomiPolicyDir() string {
	return filepath.Join(GetAptomiBaseDir(), "policy")
}

// GetAptomiObjectFilePatternYaml returns file pattern for aptomi objects (so they can be loaded from those files)
func GetAptomiObjectFilePatternYaml(baseDir string, aptomiObject AptomiObject) string {
	return filepath.Join(baseDir, "**", string(aptomiObject)+"*.yaml")
}

// GetAptomiObjectFilePatternTgz returns file pattern for tgz objects (so they can be loaded from those files)
func GetAptomiObjectFilePatternTgz(baseDir string, aptomiObject AptomiObject, chartName string) string {
	return filepath.Join(baseDir, "**", chartName+".tgz")
}
