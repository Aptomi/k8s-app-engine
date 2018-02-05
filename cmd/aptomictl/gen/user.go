package gen

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newUserCommand(cfg *config.Client) *cobra.Command {
	var username, password string

	cmd := &cobra.Command{
		Use:   "user",
		Short: "gen user",
		Long:  "gen user long",

		Run: func(cmd *cobra.Command, args []string) {
			if len(username) <= 0 {
				panic(fmt.Sprintf("username should be specified"))
			}
			if len(password) <= 0 {
				panic(fmt.Sprintf("password should be specified"))
			}

			user := &lang.User{
				Name:         username,
				PasswordHash: util.HashAndSalt(password),
			}

			data, err := yaml.Marshal(user)
			if err != nil {
				panic(fmt.Sprintf("error while marshaling user: %s", err))
			}

			fmt.Println(string(data))
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
