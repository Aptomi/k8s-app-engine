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
	var userName, userPassword string

	cmd := &cobra.Command{
		Use:   "user",
		Short: "gen user",
		Long:  "gen user long",

		Run: func(cmd *cobra.Command, args []string) {
			if len(userName) <= 0 {
				panic(fmt.Sprintf("username should be specified"))
			}
			if len(userPassword) <= 0 {
				panic(fmt.Sprintf("password should be specified"))
			}

			user := &lang.User{
				Name:         userName,
				PasswordHash: util.HashAndSalt(userPassword),
			}

			data, err := yaml.Marshal(user)
			if err != nil {
				panic(fmt.Sprintf("error while marshaling user: %s", err))
			}

			fmt.Println(string(data))
		},
	}

	cmd.Flags().StringVarP(&userName, "name", "", "", "Username")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&userPassword, "pass", "", "", "Password")
	cmd.MarkFlagRequired("pass")

	return cmd
}
