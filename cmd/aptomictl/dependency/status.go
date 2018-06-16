package dependency

import (
	"fmt"
	"time"

	"github.com/Aptomi/aptomi/cmd/aptomictl/io"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client/rest"
	"github.com/Aptomi/aptomi/pkg/client/rest/http"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/Aptomi/aptomi/pkg/util/retry"
	"github.com/gosuri/uilive"
	"github.com/gosuri/uitable"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newStatusCommand(cfg *config.Client) *cobra.Command {
	paths := make([]string, 0)
	var wait bool
	var waitInterval time.Duration
	var waitTime time.Duration
	var waitFlag string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "dependency status",
		Long:  "dependency status",

		Run: func(cmd *cobra.Command, args []string) {
			allObjects, err := io.ReadLangObjects(paths)
			if err != nil {
				panic(fmt.Sprintf("error while reading policy files: %s", err))
			}

			dependencies := []*lang.Dependency{}
			for _, obj := range allObjects {
				if d, ok := obj.(*lang.Dependency); ok {
					dependencies = append(dependencies, d)
				}
			}

			// start live updates
			writer := uilive.New()
			writer.Start()

			// if result is non-nil, then one of the dependencies is not in good state
			var result error

			if !wait {
				// query dependency status and print a one-time table with current results
				_, result = printStatusOfDependencies(cfg, dependencies, api.DependencyQueryFlag(waitFlag), writer, -1)
			} else {
				// print live updates until dependencies are ready or timeout happens
				attempt := 0
				ok := retry.Do(waitTime, waitInterval, func() bool {
					var keepWaiting bool
					keepWaiting, result = printStatusOfDependencies(cfg, dependencies, api.DependencyQueryFlag(waitFlag), writer, attempt) // nolint: gas
					attempt++
					return !keepWaiting
				})

				// if they are still not ready, let's print final results one more time (replacing progress indicators with "no")
				if !ok {
					_, result = printStatusOfDependencies(cfg, dependencies, api.DependencyQueryFlag(waitFlag), writer, -1)
				}
			}

			// stop live updates
			writer.Stop()

			// if one of the dependencies is not in good state, we should report a non-zero exit code
			if result != nil {
				log.Fatalf("one or more dependencies didn't reach '%s' state", waitFlag)
			}
		},
	}

	cmd.Flags().StringSliceVarP(&paths, "policyPaths", "f", make([]string, 0), "Paths to files/dirs with dependency files")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until dependency gets deployed and/or becomes ready. See wait-status flag")
	cmd.Flags().StringVar(&waitFlag, "state", string(api.DependencyQueryDeploymentStatusOnly),
		fmt.Sprintf("If set to '%s', the client will query dependency deployment status. If set to '%s', the client query dependency readiness status (all health checks passing)", api.DependencyQueryDeploymentStatusOnly, api.DependencyQueryDeploymentStatusAndReadiness),
	)
	cmd.Flags().DurationVar(&waitInterval, "wait-interval", 2*time.Second, "Seconds to sleep between wait attempts")
	cmd.Flags().DurationVar(&waitTime, "wait-time", 10*time.Minute, "Max time to wait before failing the wait process")

	return cmd
}

// TODO: ideally we should use common.Format() here to support writing into json and yaml, but runtime.Displayable() doesn't blend too well with an external state (i.e. dKey, waitFlag, attempt) as well as maps and sorted keys
func printStatusOfDependencies(cfg *config.Client, dependencies []*lang.Dependency, waitFlag api.DependencyQueryFlag, writer *uilive.Writer, attempt int) (bool, error) { // nolint: interfacer
	result, errAPI := rest.New(cfg, http.NewClient(cfg)).Dependency().Status(dependencies, waitFlag)
	if errAPI != nil {
		panic(fmt.Sprintf("error while requesting dependency status: %s", errAPI))
	}

	table := uitable.New()
	table.MaxColWidth = 120
	table.Wrap = true
	table.AddRow(getHeader(waitFlag)...)

	keepWaiting := false
	var err error
	for _, dKey := range util.GetSortedStringKeys(result.Status) {
		table.AddRow(getRow(dKey, result.Status[dKey], waitFlag, attempt)...)
		keepWaitingItem, errItem := shouldKeepWaiting(result.Status[dKey], waitFlag)
		if keepWaitingItem {
			keepWaiting = true
		}
		if errItem != nil {
			err = errItem
		}
	}
	fmt.Fprint(writer, table, "\n")
	return keepWaiting, err
}

func getHeader(waitFlag api.DependencyQueryFlag) []interface{} {
	result := []interface{}{"DEPENDENCY", "FOUND", "DEPLOYED"}
	if waitFlag == api.DependencyQueryDeploymentStatusAndReadiness {
		result = append(result, "READY")
	}
	return result
}

func getRow(dKey string, dStatus *api.DependencyStatus, waitFlag api.DependencyQueryFlag, attempt int) []interface{} {
	result := []interface{}{dKey, getFoundStr(dStatus), getDeployedStr(dStatus, attempt)}
	if waitFlag == api.DependencyQueryDeploymentStatusAndReadiness {
		result = append(result, getReadyStr(dStatus, attempt))
	}
	return result
}

const spinner = "|/-\\"

func getFoundStr(dsi *api.DependencyStatus) string {
	if !dsi.Found {
		return "no"
	}
	return "yes" // nolint: goconst
}

func getDeployedStr(dsi *api.DependencyStatus, attempt int) string {
	if !dsi.Found {
		return "no"
	}
	if !dsi.Deployed {
		if attempt >= 0 {
			return string(spinner[attempt%len(spinner)])
		}
		return "no"
	}
	return "yes" // nolint: goconst
}

func getReadyStr(dsi *api.DependencyStatus, attempt int) string {
	if !dsi.Found {
		return "no"
	}
	if !dsi.Ready {
		if attempt >= 0 {
			return string(spinner[attempt%len(spinner)])
		}
		return "no"
	}
	return "yes" // nolint: goconst
}

func shouldKeepWaiting(dsi *api.DependencyStatus, waitFlag api.DependencyQueryFlag) (bool, error) {
	if !dsi.Found {
		// if dependency has not been found, it does NOT make sense to continue waiting
		return false, fmt.Errorf("dependency has not been found")
	}
	if !dsi.Deployed {
		// if dependency has not been deployed (i.e. still has pending actions), we should continue waiting
		return true, fmt.Errorf("dependency is not in deployed state")
	}
	if waitFlag == api.DependencyQueryDeploymentStatusAndReadiness {
		// if dependency is not ready (i.e. health checks are still not passing), we should continue waiting
		if !dsi.Ready {
			return true, fmt.Errorf("dependency is not in ready state")
		}
	}
	// everything is ready, need to stop waiting
	return false, nil
}
