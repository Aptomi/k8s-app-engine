package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func AddStringFlag(command *cobra.Command, key, flagName, flagShorthand, defaultValue, env, usage string) {
	command.PersistentFlags().StringP(flagName, flagShorthand, defaultValue, usage)
	bindFlagEnv(command, key, flagName, env)
}

func AddBoolFlag(command *cobra.Command, key, flagName, flagShorthand string, defaultValue bool, env, usage string) {
	command.PersistentFlags().BoolP(flagName, flagShorthand, defaultValue, usage)
	bindFlagEnv(command, key, flagName, env)
}

func bindFlagEnv(command *cobra.Command, key, flagName, env string) {
	err := viper.BindPFlag(key, command.PersistentFlags().Lookup(flagName))
	if err != nil {
		log.Panicf("Error while binding flag with key %s: %s", key, err)
	}
	if len(env) > 0 {
		err = viper.BindEnv(key, env)
		if err != nil {
			log.Panicf("Error while binding env var with key %s: %s", key, err)
		}
	}
}

func AddDefaultFlags(command *cobra.Command, envPrefix string) {
	AddStringFlag(command, "config", "config", "c", "", envPrefix+"_CONFIG", "Config file or config dir path")

	AddBoolFlag(command, "debug", "debug", "d", false, envPrefix+"_DEBUG", "Debug level")

	AddStringFlag(command, "api.schema", "schema", "", "http", envPrefix+"_SCHEMA", "Server API schema")
	AddStringFlag(command, "api.host", "host", "", "127.0.0.1", envPrefix+"_HOST", "Server API host")
	AddStringFlag(command, "api.port", "port", "", "27866", envPrefix+"_PORT", "Server API port")
	AddStringFlag(command, "api.apiPrefix", "api-prefix", "", "api/v1", envPrefix+"_API_PREFIX", "Server API prefix")
}
