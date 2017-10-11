package main

import (
	"github.com/Aptomi/aptomi/cmd"
	"github.com/Aptomi/aptomi/pkg/config"
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
		Use:   "aptomictl",
		Short: "aptomictl controls Aptomi",
		Long:  "aptomictl controls Aptomi",

		PersistentPreRun: preRun,

		Run: func(cmd *cobra.Command, args []string) {
			// fall back on default help if no args/flags are passed
			cmd.HelpFunc()(cmd, args)
		},
	}
)

func init() {
	viper.SetEnvPrefix(envPrefix)

	cmd.AddDefaultFlags(aptomiCtlCmd, envPrefix)

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
