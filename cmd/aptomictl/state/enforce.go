package state

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/aptomictl/util"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

func newEnforceCommand(cfg *config.Client) *cobra.Command {
	var wait bool
	var noop bool
	var waitInterval time.Duration
	var waitAttempts int

	cmd := &cobra.Command{
		Use:   "enforce",
		Short: "state enforce",
		Long:  "state enforce long",

		Run: func(cmd *cobra.Command, args []string) {
			// call API (apply or delete), get policy update result
			clientObj := rest.New(cfg, http.NewClient(cfg))
			result, err := clientObj.State().Reset(noop)
			if err != nil {
				panic(fmt.Sprintf("error while calling state reset: %s", err))
			}

			// print policy update result to the screen
			util.PrintPolicyUpdateResult(result, logrus.WarnLevel, cfg)

			// wait for actions to finish, if needed
			if wait {
				util.WaitForRevisionActionsToFinish(waitAttempts, waitInterval, clientObj, result)
			}

		},
	}

	cmd.Flags().BoolVar(&noop, "noop", false, "Produce action plan for the given changes in policy, but do not run any actions to update the state")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until all actions are fully applied")
	cmd.Flags().DurationVar(&waitInterval, "wait-interval", 2*time.Second, "Seconds to sleep between wait attempts")
	cmd.Flags().IntVar(&waitAttempts, "wait-attempts", 150, "Number of wait attempts before failing the wait process")

	return cmd
}
