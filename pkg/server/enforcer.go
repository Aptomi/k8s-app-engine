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

	// Mark last Revision as failed if it wasn't completed
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

	// TODO: get rid of err revisions. revision should not have a gloval err status
	if false {
		server.saveErrRevision(currRevision, desiredPolicyGen, resolveLog)
		return fmt.Errorf("cannot resolve desiredPolicy: %s", err)
	}

	stateDiff := diff.NewPolicyResolutionDiff(desiredState, actualState)

	nextRevision, err := server.store.NewRevision(desiredPolicyGen)
	if err != nil {
		return fmt.Errorf("unable to get next revision: %s", err)
	}
	nextRevision.ResolveLog = resolveLog.AsAPIEvents()

	// policy changed while no actions needed to achieve desired state
	if len(stateDiff.Actions) <= 0 && currRevision != nil && currRevision.Policy == nextRevision.Policy {
		log.Infof("(enforce-%d) No changes, policy gen %d", server.enforcementIdx, desiredPolicyGen)
		return nil
	}
	log.Infof("(enforce-%d) New revision %d, policy gen %d, %d actions need to be applied", server.enforcementIdx, nextRevision.GetGeneration(), desiredPolicyGen, len(stateDiff.Actions))

	// Save revision
	err = server.store.SaveRevision(nextRevision)
	if err != nil {
		return fmt.Errorf("error while saving new revision: %s", err)
	}

	if server.cfg.Enforcer.Noop {
		log.Infof("(enforce-%d) Applying changes in noop mode (sleep per action = %s)", server.enforcementIdx, server.cfg.Enforcer.NoopSleep)
	} else {
		log.Infof("(enforce-%d) Applying changes", server.enforcementIdx)
	}

	pluginRegistry := server.pluginRegistryFactory()
	applyLog := event.NewLog(fmt.Sprintf("enforce-%d-apply", server.enforcementIdx), true)
	applier := apply.NewEngineApply(desiredPolicy, desiredState, actualState, server.store.GetActualStateUpdater(), server.externalData, pluginRegistry, stateDiff.Actions, applyLog, server.store.GetRevisionProgressUpdater(nextRevision))
	_, err = applier.Apply()

	// reload revision to have progress data saved into it
	nextRevision, saveErr := server.store.GetRevision(runtime.LastGen)
	if saveErr != nil {
		return fmt.Errorf("error while reloading last revision to have progress loaded: %s", saveErr)
	}
	nextRevision.ApplyLog = applyLog.AsAPIEvents()

	// save apply log
	saveErr = server.store.UpdateRevision(nextRevision)
	if saveErr != nil {
		return fmt.Errorf("error while saving new revision with apply log: %s", saveErr)
	}

	if err != nil {
		return fmt.Errorf("error while applying new revision: %s", err)
	}
	log.Infof("(enforce-%d) New revision %d successfully applied, %d component instances", server.enforcementIdx, nextRevision.GetGeneration(), len(desiredState.GetComponentProcessingOrder()))

	return nil
}

func (server *Server) saveErrRevision(currRevision *engine.Revision, desiredPolicyGen runtime.Generation, resolveLog *event.Log) {
	if currRevision == nil || currRevision.Policy != desiredPolicyGen || currRevision.Status != engine.RevisionStatusError {
		rev, err := server.store.NewRevision(desiredPolicyGen)
		if err != nil {
			log.Warnf("(enforce-%d) Error while creating revision to record resolution error: %s", server.enforcementIdx, err)
		}

		rev.Status = engine.RevisionStatusError
		rev.ResolveLog = resolveLog.AsAPIEvents()

		err = server.store.SaveRevision(rev)
		if err != nil {
			log.Warnf("(enforce-%d) Error while saving revision to record resolution error: %s", server.enforcementIdx, err)
		}
	}
}
