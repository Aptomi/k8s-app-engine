package login

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns instance of cobra command that allows to login into aptomi
func NewCommand(cfg *config.Client) *cobra.Command {
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

			fmt.Println("New token:", authSuccess.Token)
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
