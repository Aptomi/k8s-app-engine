package main

import (
	"github.com/Aptomi/aptomi/cmd"
	"github.com/Aptomi/aptomi/pkg/slinga/config"
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path"
)

const (
	envPrefix = "APTOMICTL"
)

var (
	cfg          = &config.Client{}
	aptomiCtlCmd = &cobra.Command{
		Use:   "aptomictl", // todo(slukjanov)
		Short: "",          // todo(slukjanov)
		Long:  "",          // todo(slukjanov)

		// parse the configPath if one is provided, or use the defaults
		PersistentPreRun: preRun,

		Run: func(cmd *cobra.Command, args []string) {
			// fall back on default help if no args/flags are passed
			cmd.HelpFunc()(cmd, args)
		},
	}
)

func init() {
	viper.SetEnvPrefix(envPrefix)

	aptomiCtlCmd.PersistentFlags().StringP("config", "c", "", "Config file or dir path")

	aptomiCtlCmd.PersistentFlags().String("host", "127.0.0.1", "Server API host")
	err := viper.BindPFlag("server.host", aptomiCtlCmd.PersistentFlags().Lookup("host"))
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}
	err = viper.BindEnv("server.host", envPrefix+"_HOST")
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}

	aptomiCtlCmd.PersistentFlags().Uint16P("port", "p", 27866, "Server API port")
	err = viper.BindPFlag("server.port", aptomiCtlCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}
	err = viper.BindEnv("server.port", envPrefix+"_PORT")
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}

	aptomiCtlCmd.AddCommand(cmd.Version)
}

func preRun(command *cobra.Command, args []string) {
	cmd.ReadConfig(viper.GetViper(), cfg, defaultConfigDir())
}

func defaultConfigDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Panicf("Can't find homedir: %s", err)
	}

	return path.Join(home, ".aptomi")
}
