package login

import (
	"bytes"
	"fmt"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// NewCommand returns instance of cobra command that allows to login into aptomi
func NewCommand(cfg *config.Client, cfgFile *string) *cobra.Command {
	var username, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login into the Aptomi",
		Run: func(cmd *cobra.Command, args []string) {
			if len(username) == 0 || len(password) == 0 {
				log.Fatalf("username and password should not be both empty")
			}

			authSuccess, err := rest.New(cfg, http.NewClient(cfg)).User().Login(username, password)
			if err != nil {
				log.Fatalf("error while user login: %s", err)
			}

			cfg.Auth.Token = authSuccess.Token

			writeConfig(cfg, cfgFile)

			log.Infof("Config successfully updated with token")
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	if err := cmd.MarkFlagRequired("username"); err != nil {
		panic(err)
	}
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	if err := cmd.MarkFlagRequired("password"); err != nil {
		panic(err)
	}

	return cmd
}

func writeConfig(cfg *config.Client, cfgFile *string) {
	cleanupDefaultsFromConfig(cfg)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		panic(fmt.Sprintf("error marshaling client config: %s", err))
	}

	backupConfigIfDiffers(*cfgFile, data)

	err = ioutil.WriteFile(*cfgFile, data, 0644)
	if err != nil {
		panic(fmt.Sprintf("error while saving config with token to %s: %s", *cfgFile, err))
	}
}

func cleanupDefaultsFromConfig(cfg *config.Client) {
	if cfg.Output == common.Default {
		cfg.Output = ""
	}
	if cfg.API.Schema == common.DefaultAPISchema {
		cfg.API.Schema = ""
	}
	// Keeping API and Port in the config to make it easier to customize
	/*
		if cfg.API.Host == common.DefaultAPIHost {
			cfg.API.Host = ""
		}
		if cfg.API.Port == common.DefaultAPIPort {
			cfg.API.Port = 0
		}
	*/
	if cfg.API.APIPrefix == common.DefaultAPIPrefix {
		cfg.API.APIPrefix = ""
	}
}

func backupConfigIfDiffers(cfgFile string, newData []byte) {
	cfg, err := os.Open(cfgFile)
	if err != nil {
		panic(fmt.Sprintf("error while opening current config file %s: %s", cfgFile, err))
	}
	defer func() {
		closeErr := cfg.Close()
		if closeErr != nil {
			panic(fmt.Sprintf("error while closing current config file %s: %s", cfgFile, err))
		}
	}()

	data, err := ioutil.ReadAll(cfg)
	if err != nil {
		panic(fmt.Sprintf("error while reading current config file %s: %s", cfgFile, err))
	}

	if !bytes.Equal(data, newData) {
		backupCfgFile := cfgFile + ".bak"
		err = ioutil.WriteFile(backupCfgFile, data, 0644)
		if err != nil {
			panic(fmt.Sprintf("error while writing backup of the current config %s: %s", cfgFile, err))
		}
		log.Infof("Current config saved to: %s", backupCfgFile)
	}
}
