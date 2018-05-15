package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/apply"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/runtime"
	log "github.com/Sirupsen/logrus"
	"time"
)

func logError(err interface{}) {
	log.Errorf("Error while enforcing policy: %s", err)
}

func (server *Server) enforceLoop() error {
	for {
		err := server.enforce()
		if err != nil {
			logError(err)
		}

		// sleep for a specified time or wait until policy has changed, whichever comes first
		timer := time.NewTimer(server.cfg.Enforcer.Interval)
		select {
		case <-server.policyChanged:
			break // nolint: megacheck
		case <-timer.C:
			break // nolint: megacheck
		}
		timer.Stop()
	}
}

func (server *Server) enforce() error {
	server.enforcementIdx++

	defer func() {
		if err := recover(); err != nil {
			logError(err)
		}
	}()

	// todo think about initial state when there is no revision at all
	currRevision, err := server.store.GetRevision(runtime.LastGen)
	if err != nil {
		return fmt.Errorf("unable to get curr revision: %s", err)
	}

	// Mark last Revision as failed if it wasn't completed. If it's in 'waiting' state, that's probably ok
	if currRevision != nil && currRevision.Status == engine.RevisionStatusInProgress {
		currRevision.Status = engine.RevisionStatusError
		currRevision.AppliedAt = time.Now()
		revErr := server.store.UpdateRevision(currRevision)
		if revErr != nil {
			log.Warnf("(enforce-%d) Error while setting current revision that is in progress to error state: %s", server.enforcementIdx, revErr)
		}
		log.Infof("(enforce-%d) Current revision that is in progress was reset to error state", server.enforcementIdx)
	}

	desiredPolicy, desiredPolicyGen, err := server.store.GetPolicy(runtime.LastGen)
	if err != nil {
		return fmt.Errorf("error while getting desiredPolicy: %s", err)
	}

	// if policy is not found, it means it somehow was not initialized correctly. let's return error
	if desiredPolicy == nil {
		return fmt.Errorf("desiredPolicy is nil, does not exist in the store")
	}

	actualState, err := server.store.GetActualState()
	if err != nil {
		return fmt.Errorf("error while getting actual state: %s", err)
	}

	resolveLog := event.NewLog(fmt.Sprintf("enforce-%d-resolve", server.enforcementIdx), true)
	resolver := resolve.NewPolicyResolver(desiredPolicy, server.externalData, resolveLog)
	desiredState := resolver.ResolveAllDependencies()

	stateDiff := diff.NewPolicyResolutionDiff(desiredState, actualState)

	nextRevision, err := server.store.NewRevision(desiredPolicyGen)
	if err != nil {
		return fmt.Errorf("unable to get next revision: %s", err)
	}
	nextRevision.ResolveLog = resolveLog.AsAPIEvents()

	// policy changes while no actions needed to achieve desired state
	actionCnt := stateDiff.ActionPlan.NumberOfActions()
	if actionCnt <= 0 && currRevision != nil && currRevision.Policy == nextRevision.Policy {
		log.Infof("(enforce-%d) No changes, policy gen %d", server.enforcementIdx, desiredPolicyGen)
		return nil
	}
	log.Infof("(enforce-%d) New revision %d, policy gen %d, %d actions need to be applied", server.enforcementIdx, nextRevision.GetGeneration(), desiredPolicyGen, actionCnt)

	// Save revision
	err = server.store.SaveRevision(nextRevision)
	if err != nil {
		return fmt.Errorf("error while saving new revision: %s", err)
	}

	if server.cfg.Enforcer.Noop {
		log.Infof("(enforce-%d) Applying actions in noop mode (sleep per action = %s)", server.enforcementIdx, server.cfg.Enforcer.NoopSleep)
	} else {
		log.Infof("(enforce-%d) Applying actions", server.enforcementIdx)
	}

	pluginRegistry := server.pluginRegistryFactory()
	applyLog := event.NewLog(fmt.Sprintf("enforce-%d-apply", server.enforcementIdx), true)
	applier := apply.NewEngineApply(desiredPolicy, desiredState, actualState, server.store.GetActualStateUpdater(), server.externalData, pluginRegistry, stateDiff.ActionPlan, applyLog, server.store.NewRevisionResultUpdater(nextRevision))
	_, _ = applier.Apply()

	// save apply log
	nextRevision.ApplyLog = applyLog.AsAPIEvents()
	saveErr := server.store.UpdateRevision(nextRevision)
	if saveErr != nil {
		return fmt.Errorf("error while saving new revision with apply log: %s", saveErr)
	}

	log.Infof("(enforce-%d) New revision %d processed, %d component instances", server.enforcementIdx, nextRevision.GetGeneration(), len(desiredState.ComponentInstanceMap))

	return nil
}
