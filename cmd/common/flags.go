package common

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

// AddStringFlag adds string flag to provided cobra command and registers with provided env variable name
func AddStringFlag(command *cobra.Command, key, flagName, flagShorthand, defaultValue, env, usage string) {
	command.PersistentFlags().StringP(flagName, flagShorthand, defaultValue, usage)
	bindFlagEnv(command, key, flagName, env)
}

// AddBoolFlag adds bool flag to provided cobra command and registers with provided env variable name
func AddBoolFlag(command *cobra.Command, key, flagName, flagShorthand string, defaultValue bool, env, usage string) {
	command.PersistentFlags().BoolP(flagName, flagShorthand, defaultValue, usage)
	bindFlagEnv(command, key, flagName, env)
}

// AddDurationFlag adds duration flag to provided cobra command and registers with provided env variable name
func AddDurationFlag(command *cobra.Command, key, flagName, flagShorthand string, defaultValue time.Duration, env, usage string) {
	command.PersistentFlags().DurationP(flagName, flagShorthand, defaultValue, usage)
	bindFlagEnv(command, key, flagName, env)
}

func bindFlagEnv(command *cobra.Command, key, flagName, env string) {
	err := viper.BindPFlag(key, command.PersistentFlags().Lookup(flagName))
	if err != nil {
		panic(fmt.Sprintf("Error while binding flag with key %s: %s", key, err))
	}
	if len(env) > 0 {
		err = viper.BindEnv(key, env)
		if err != nil {
			panic(fmt.Sprintf("Error while binding env var with key %s: %s", key, err))
		}
	}
}

const (
	// DefaultApiSchema is a default API schema
	DefaultApiSchema = "http"
	// DefaultApiHost is a default API host
	DefaultApiHost = "127.0.0.1"
	// DefaultApiPort is a default API port
	DefaultApiPort = 27866
	// DefaultApiPrefix is a default API prefix
	DefaultApiPrefix = "api/v1"
)

// AddDefaultFlags add all the flags that are needed by any aptomi CLI
func AddDefaultFlags(command *cobra.Command, envPrefix string) {
	AddStringFlag(command, "config", "config", "", "", envPrefix+"_CONFIG", "Config file or config dir path")

	AddBoolFlag(command, "debug", "debug", "d", false, envPrefix+"_DEBUG", "Debug level")

	AddStringFlag(command, "api.schema", "api-schema", "", DefaultApiSchema, envPrefix+"_SCHEMA", "Server API schema")
	AddStringFlag(command, "api.host", "api-host", "", DefaultApiHost, envPrefix+"_HOST", "Server API host")
	AddStringFlag(command, "api.port", "api-port", "", strconv.Itoa(DefaultApiPort), envPrefix+"_PORT", "Server API port")
	AddStringFlag(command, "api.apiPrefix", "api-prefix", "", DefaultApiPrefix, envPrefix+"_API_PREFIX", "Server API prefix")
}
