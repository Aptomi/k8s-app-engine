package server

import (
	"fmt"
	"time"

	"runtime/debug"

	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/apply"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func (server *Server) desiredStateEnforceLoop() error {
	server.desiredStateEnforcements = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name:        "aptomi_desired_state_enforcements_total",
			Help:        "Total number of completed desired state enforcements",
			ConstLabels: prometheus.Labels{"service": prometheusSvcName},
		},
	)
	prometheus.MustRegister(server.desiredStateEnforcements)

	// todo consider converting into histogram vector and labeling with stage (no rev, no changes), policy and rev gens
	server.desiredStateEnforcementDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "aptomi_desired_state_enforcement_duration_seconds",
		Help:        "Duration of the completed desired state enforcements",
		ConstLabels: prometheus.Labels{"service": prometheusSvcName},
		Buckets:     []float64{.1, 1, 10, 20, 30, 60, 120, 180, 300, 600},
	},
	)
	prometheus.MustRegister(server.desiredStateEnforcementDuration)

	for {
		err := server.desiredStateEnforce()
		if err != nil {
			log.Errorf("error while enforcing desired state: %s", err)
		}

		// sleep for a specified time or wait until policy has changed, whichever comes first
		timer := time.NewTimer(server.cfg.Enforcer.Interval)
		select {
		case <-server.runDesiredStateEnforcement:
			break // nolint: megacheck
		case <-timer.C:
			break // nolint: megacheck
		}
		timer.Stop()
	}
}

func (server *Server) getRevisionForProcessing() (*engine.Revision, error) {
	// we are processing revision sequentially, so let's get the first unprocessed revision from the database
	revision, err := server.store.GetFirstUnprocessedRevision()
	if err != nil {
		return nil, fmt.Errorf("unable to load first unprocessed revision: %s", err)
	}

	// if there is an unprocessed revision, return it
	if revision != nil {
		log.Infof("(enforce-%d) Found unprocessed revision %d", server.desiredStateEnforcementIdx, revision.GetGeneration())
		return revision, nil
	}

	// if there are no unprocessed revisions, let's get the last one and see if it was successful or not
	_, policyGen, err := server.store.GetPolicy(runtime.LastGen)
	if err != nil {
		return nil, fmt.Errorf("unable to load latest policy: %s", err)
	}
	lastRevision, err := server.store.GetLastRevisionForPolicy(policyGen)
	if err != nil {
		return nil, fmt.Errorf("unable to load latest revision: %s", err)
	}

	// now, given that we retrieved the last revision, when do we need to retry it? in one of two cases:
	// - it's either in error status (something really bad happened)
	// - it completed, but some actions failed and they need to be retried
	if lastRevision != nil && (lastRevision.Status == engine.RevisionStatusError || (lastRevision.Status == engine.RevisionStatusCompleted && lastRevision.Result.Failed > 0)) {
		log.Infof("(enforce-%d) Found last revision %d which needs to be retried", server.desiredStateEnforcementIdx, lastRevision.GetGeneration())
		return lastRevision, nil
	}

	// nothing to process
	return nil, nil
}

func (server *Server) desiredStateEnforce() error {
	start := time.Now()
	server.desiredStateEnforcementIdx++

	defer func() {
		server.desiredStateEnforcements.Inc()
		server.desiredStateEnforcementDuration.Observe(time.Since(start).Seconds())

		if err := recover(); err != nil {
			log.Errorf("panic while enforcing desired state: %s", err)
			log.Errorf(string(debug.Stack()))
		}
	}()

	// get the revision for processing
	revision, err := server.getRevisionForProcessing()
	if err != nil {
		return fmt.Errorf("can't pick revision for processing: %s", err)
	}
	if revision == nil {
		return nil
	}

	// reset revision status and result
	revision.Status = engine.RevisionStatusWaiting
	revision.Result = &action.ApplyResult{}
	revErr := server.store.UpdateRevision(revision)
	if revErr != nil {
		return fmt.Errorf("unable to update revision: %s", revErr)
	}

	// load the corresponding policy
	policy, policyGen, err := server.store.GetPolicy(revision.PolicyGen)
	if err != nil {
		return fmt.Errorf("error while getting policy: %s", err)
	}

	// load desired state
	desiredState, err := server.store.GetDesiredState(revision)
	if err != nil {
		return fmt.Errorf("can't load desired state from revision: %s", err)
	}

	// load the actual state
	actualState, err := server.store.GetActualState()
	if err != nil {
		return fmt.Errorf("error while getting actual state: %s", err)
	}

	// compare desired against actual
	var stateDiff *diff.PolicyResolutionDiff
	if revision.RecalculateAll {
		stateDiff = diff.NewPolicyResolutionDiff(desiredState, resolve.NewPolicyResolution())
	} else {
		stateDiff = diff.NewPolicyResolutionDiff(desiredState, actualState)
	}

	// policy changes while no actions needed to achieve desired state
	actionCnt := stateDiff.ActionPlan.NumberOfActions()
	if actionCnt > 0 {
		log.Infof("(enforce-%d) Revision %d, policy gen %d: %d actions need to be applied", server.desiredStateEnforcementIdx, revision.GetGeneration(), policyGen, actionCnt)
	} else {
		log.Infof("(enforce-%d) Revision %d, policy gen %d: no changes", server.desiredStateEnforcementIdx, revision.GetGeneration(), policyGen)
	}

	// op or noop
	if server.cfg.Enforcer.Noop {
		log.Infof("(enforce-%d) Applying actions in noop mode (sleep per action = %s)", server.desiredStateEnforcementIdx, server.cfg.Enforcer.NoopSleep)
	} else {
		log.Infof("(enforce-%d) Applying actions", server.desiredStateEnforcementIdx)
	}

	// apply
	pluginRegistry := server.enforcerPluginRegistryFactory()
	applyLog := event.NewLog(log.DebugLevel, fmt.Sprintf("enforce-%d-apply", server.desiredStateEnforcementIdx)).AddConsoleHook(server.cfg.GetLogLevel())
	applier := apply.NewEngineApply(policy, desiredState, server.store.NewActualStateUpdater(actualState), server.externalData, pluginRegistry, stateDiff.ActionPlan, applyLog, server.store.NewRevisionResultUpdater(revision))
	_, _ = applier.Apply(server.cfg.Enforcer.MaxConcurrentActions)

	// save apply log
	revision.ApplyLog = applyLog.AsAPIEvents()
	saveErr := server.store.UpdateRevision(revision)
	if saveErr != nil {
		return fmt.Errorf("error while saving revision with apply log: %s", saveErr)
	}

	log.Infof("(enforce-%d) Revision %d processed (actions: %d succeeded, %d failed, %d skipped)", server.desiredStateEnforcementIdx, revision.GetGeneration(), revision.Result.Success, revision.Result.Failed, revision.Result.Skipped)

	// let's try again immediately until no actions were successfully applied
	if revision.Result.Success > 0 {
		// trigger enforcement again
		server.runDesiredStateEnforcement <- true
		// trigger actual state update
		server.runActualStateUpdate <- true
	}

	return nil
}
