package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/apply"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	log "github.com/sirupsen/logrus"
	"time"
)

func (server *Server) desiredStateEnforceLoop() error {
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

func (server *Server) desiredStateEnforce() error {
	server.desiredStateEnforcementIdx++

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic while enforcing desired state: %s", err)
		}
	}()

	// load revision from the head of the queue for processing
	revision, err := server.store.GetFirstUnprocessedRevision()
	if err != nil {
		return fmt.Errorf("unable to get first unprocessed revision: %s", err)
	}

	// if there is no revision, we have no work to do
	if revision == nil {
		return nil
	}

	// it this revision has been stuck before, we will need to retry it
	if revision.Status != engine.RevisionStatusWaiting {
		revision.Status = engine.RevisionStatusWaiting
		revErr := server.store.UpdateRevision(revision)
		if revErr != nil {
			return fmt.Errorf("unable to save revision: %s", err)
		}
		log.Infof("(enforce-%d) took revision %d from the queue, but it's in progress. resetting to waiting state and reapplying changes", server.desiredStateEnforcementIdx)
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

	log.Infof("(enforce-%d) Revision %d processed, %d component instances", server.desiredStateEnforcementIdx, revision.GetGeneration(), len(desiredState.ComponentInstanceMap))

	// trigger enforcement again
	server.runDesiredStateEnforcement <- true

	// trigger actual state update
	server.runActualStateUpdate <- true

	return nil
}
