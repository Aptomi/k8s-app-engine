package login

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// NewCommand returns instance of cobra command that allows to login into aptomi
func NewCommand(cfg *config.Client, cfgFile *string) *cobra.Command {
	var username, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "login into the Aptomi",
		Run: func(cmd *cobra.Command, args []string) {
			if len(username) == 0 || len(password) == 0 {
				panic(fmt.Sprintf("Both username and password should be non-empty"))
			}

			authSuccess, err := rest.New(cfg, http.NewClient(cfg)).User().Login(username, password)
			if err != nil {
				panic(fmt.Sprintf("Error while user login: %s", err))
			}

			cfg.Auth.Token = authSuccess.Token

			data, err := yaml.Marshal(cfg)
			if err != nil {
				panic(fmt.Sprintf("Error marshaling client config: %s", err))
			}
			err = ioutil.WriteFile(*cfgFile, data, 0644)
			if err != nil {
				panic(fmt.Sprintf("Error while saving config with token to %s: %s", *cfgFile, err))
			}

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
