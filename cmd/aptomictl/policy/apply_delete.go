package policy

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/aptomictl/io"
	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/progress"
	"github.com/Aptomi/aptomi/pkg/util/retry"
	"github.com/Sirupsen/logrus"
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
				panic(fmt.Sprintf("error while reading policy files: %s", err))
			}

			logLevelObj, err := logrus.ParseLevel(logLevel)
			if err != nil {
				logLevelObj = logrus.WarnLevel
			}

			clientObj := rest.New(cfg, http.NewClient(cfg))
			var result *api.PolicyUpdateResult
			if createUpdate {
				result, err = clientObj.Policy().Apply(allObjects, noop, logLevelObj)
			} else {
				result, err = clientObj.Policy().Delete(allObjects, noop, logLevelObj)
			}
			if err != nil {
				panic(fmt.Sprintf("error while calling %s on policy: %s", commandType, err))
			}

			fmt.Printf("Event Log (>%s):\n", logLevelObj.String())
			if len(result.EventLog) > 0 {
				for _, entry := range result.EventLog {
					fmt.Printf("[%s] %s\n", entry.LogLevel, entry.Message)
				}
			} else {
				fmt.Println("* no entries")
			}

			data, err := common.Format(cfg.Output, false, result)
			if err != nil {
				panic(fmt.Sprintf("error while formating policy update result: %s", err))
			}
			fmt.Println(string(data))

			if !wait {
				return
			}

			waitForActionsToFinish(waitAttempts, waitInterval, clientObj, result)
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
	cmd.Flags().StringVar(&logLevel, "log-level", logrus.WarnLevel.String(), fmt.Sprintf("Retrieve logs from the server using the specified log level (%s)", logrus.AllLevels))

	return cmd
}

func waitForActionsToFinish(attempts int, interval time.Duration, clientObj client.Core, result *api.PolicyUpdateResult) {
	// if policy hasn't changed, then we don't have to wait. let's exit right away
	if !result.PolicyChanged {
		return
	}

	// wait for the revision to be processed & applied
	fmt.Print("Waiting for actions to be applied...")
	var rev *engine.Revision

	var progressBar progress.Indicator
	var progressLast = 0

	// query revision status [attempts] x [interval]
	finished := retry.Do2(attempts, interval, func() bool {
		// call API
		var revErr error
		rev, revErr = clientObj.Revision().ShowByPolicy(result.PolicyGeneration)
		if revErr != nil {
			fmt.Print(".")
			return false
		}

		// if the engine already started processing the revision, show its progress
		if rev.Status != engine.RevisionStatusWaiting {
			if progressBar == nil {
				fmt.Println()

				// show progress bar only when there is at least one action present
				if rev.Result.Total > 0 {
					progressBar = progress.NewConsole("Applying actions")
					progressBar.SetTotal(int(rev.Result.Total))
				}
			}
			for progressBar != nil && progressLast < int(rev.Result.Success+rev.Result.Failed+rev.Result.Skipped) {
				progressBar.Advance()
				progressLast++
			}
		}

		// exit when revision is in completed or error status
		return rev.Status == engine.RevisionStatusCompleted || rev.Status == engine.RevisionStatusError
	})

	// stop progress bar
	if progressBar != nil {
		progressBar.Done()
	}

	// print the outcome
	if !finished {
		fmt.Printf("Revision %d timeout! Has not been applied in %d seconds\n", rev.GetGeneration(), int(interval.Seconds()*float64(attempts)))
		panic("timeout")
	} else if rev.Status == engine.RevisionStatusCompleted {
		if rev.Result.Total > 0 {
			fmt.Printf("Revision %d completed. Actions: %d succeeded, %d failed, %d skipped\n", rev.GetGeneration(), rev.Result.Success, rev.Result.Failed, rev.Result.Skipped)
		} else {
			fmt.Printf("Revision %d completed\n", rev.GetGeneration())
		}
	} else if rev.Status == engine.RevisionStatusError {
		fmt.Printf("Revision %d failed\n", rev.GetGeneration())
		panic("error")
	} else {
		fmt.Printf("Unexpected revision status '%s' for revision %d\n", rev.Status, rev.GetGeneration())
		panic("error")
	}

}
