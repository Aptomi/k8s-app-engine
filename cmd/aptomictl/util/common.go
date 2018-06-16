package util

import (
	"fmt"
	"time"

	"github.com/Aptomi/aptomi/cmd/common"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/client"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/progress"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util/retry"
	log "github.com/sirupsen/logrus"
)

// WaitForRevisionActionsToFinish waits until revision is done (i.e. all of its pending actions are completed)
func WaitForRevisionActionsToFinish(maxTime time.Duration, interval time.Duration, clientObj client.Core, result *api.PolicyUpdateResult) {
	// if there is no revision to wait for, then exit
	if result.WaitForRevision >= runtime.MaxGeneration {
		return
	}

	// wait for the revision to be processed & applied
	fmt.Printf("Waiting for revision %d...", result.WaitForRevision)
	var rev *engine.Revision

	var progressBar progress.Indicator
	var progressLast = 0

	// query revision status
	finished := retry.Do2(maxTime, interval, func() bool {
		// call API
		var revErr error
		rev, revErr = clientObj.Revision().Show(result.WaitForRevision)
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
		log.Fatalf("Revision %d timeout! Has not been applied in %s\n", rev.GetGeneration(), maxTime)
	} else if rev.Status == engine.RevisionStatusCompleted {
		if rev.Result.Total > 0 {
			fmt.Printf("Revision %d completed. Actions: %d succeeded, %d failed, %d skipped\n", rev.GetGeneration(), rev.Result.Success, rev.Result.Failed, rev.Result.Skipped)
		} else {
			fmt.Printf("Revision %d completed\n", rev.GetGeneration())
		}
	} else if rev.Status == engine.RevisionStatusError {
		log.Fatalf("Revision %d failed\n", rev.GetGeneration())
	} else {
		log.Fatalf("Unexpected revision status '%s' for revision %d\n", rev.Status, rev.GetGeneration())
	}

}

// PrintPolicyUpdateResult prints PolicyUpdateResult to the console
func PrintPolicyUpdateResult(result *api.PolicyUpdateResult, logLevelObj log.Level, cfg *config.Client) { // nolint: interfacer
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
}
