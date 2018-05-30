package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/aptomictl/io"
	"github.com/Aptomi/aptomi/cmd/aptomictl/util"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

// if apply is true, it will apply policy changes. otherwise it will
func newHandlePolicyChangesCommand(cfg *config.Client, createUpdate bool) *cobra.Command {
	paths := make([]string, 0)
	var wait bool
	var noop bool
	var waitInterval time.Duration
	var waitAttempts int
	var logLevel string
	commandType := "apply"
	if !createUpdate {
		commandType = "delete"
	}

	cmd := &cobra.Command{
		Use:   commandType,
		Short: fmt.Sprintf("%s policy", commandType),
		Long:  fmt.Sprintf("%s policy long", commandType),

		Run: func(cmd *cobra.Command, args []string) {
			allObjects, err := io.ReadLangObjects(paths)
			if err != nil {
				log.Fatalf("error while reading policy files: %s", err)
			}

			logLevelObj, err := log.ParseLevel(logLevel)
			if err != nil {
				logLevelObj = log.WarnLevel
			}

			// call API (apply or delete), get policy update result
			clientObj := rest.New(cfg, http.NewClient(cfg))
			var result *api.PolicyUpdateResult
			if createUpdate {
				result, err = clientObj.Policy().Apply(allObjects, noop, logLevelObj)
			} else {
				result, err = clientObj.Policy().Delete(allObjects, noop, logLevelObj)
			}
			if err != nil {
				log.Fatalf("error while calling %s on policy: %s", commandType, err)
			}

			// print policy update result to the screen
			util.PrintPolicyUpdateResult(result, logLevelObj, cfg)

			// wait for actions to finish, if needed
			if wait {
				util.WaitForRevisionActionsToFinish(waitAttempts, waitInterval, clientObj, result)
			}

		},
	}

	cmd.Flags().StringSliceVarP(&paths, "policyPaths", "f", make([]string, 0), "Paths to files/dirs with policy files")
	if err := cmd.MarkFlagRequired("policyPaths"); err != nil {
		panic(err)
	}
	cmd.Flags().BoolVar(&noop, "noop", false, "Produce action plan for the given changes in policy, but do not run any actions to update the state")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until all actions are fully applied")
	cmd.Flags().DurationVar(&waitInterval, "wait-interval", 2*time.Second, "Seconds to sleep between wait attempts")
	cmd.Flags().IntVar(&waitAttempts, "wait-attempts", 150, "Number of wait attempts before failing the wait process")
	cmd.Flags().StringVar(&logLevel, "log-level", log.WarnLevel.String(), fmt.Sprintf("Retrieve logs from the server using the specified log level (%s)", log.AllLevels))

	return cmd
}
