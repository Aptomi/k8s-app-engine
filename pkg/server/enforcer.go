package server

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/apply"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/plugin/helm"
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
		time.Sleep(server.cfg.Enforcer.Interval)
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

	eventLog := event.NewLog(fmt.Sprintf("enforce-%d-resolve", server.enforcementIdx), true)
	resolver := resolve.NewPolicyResolver(desiredPolicy, server.externalData, eventLog)
	desiredState, err := resolver.ResolveAllDependencies()
	if err != nil {
		// todo save eventlog
		return fmt.Errorf("cannot resolve desiredPolicy: %s", err)
	}

	stateDiff := diff.NewPolicyResolutionDiff(desiredState, actualState)

	nextRevision, err := server.store.NewRevision(desiredPolicyGen)
	if err != nil {
		return fmt.Errorf("unable to get next revision: %s", err)
	}

	// policy changed while no actions needed to achieve desired state
	if len(stateDiff.Actions) <= 0 && currRevision != nil && currRevision.Policy == nextRevision.Policy {
		log.Infof("(enforce-%d) No changes, policy gen %d", server.enforcementIdx, desiredPolicyGen)
		return nil
	}
	log.Infof("(enforce-%d) New revision %d, policy gen %d, %d actions need to be applied", server.enforcementIdx, nextRevision.GetGeneration(), desiredPolicyGen, len(stateDiff.Actions))

	// todo save eventlog (if there were changes?)

	// todo if policy gen changed, we still need to save revision but with progress == done

	// Save revision
	err = server.store.SaveRevision(nextRevision)
	if err != nil {
		return fmt.Errorf("error while saving new revision: %s", err)
	}

	// Build plugin registry
	var pluginRegistry plugin.Registry
	if server.cfg.Enforcer.Noop {
		log.Infof("(enforce-%d) Applying changes in noop mode (sleep per action = %d seconds)", server.enforcementIdx, server.cfg.Enforcer.NoopSleep)
		pluginRegistry = &plugin.MockRegistry{
			DeployPlugin:      &plugin.MockDeployPlugin{SleepTime: time.Second * time.Duration(server.cfg.Enforcer.NoopSleep)},
			PostProcessPlugin: &plugin.MockPostProcessPlugin{},
		}
	} else {
		log.Infof("(enforce-%d) Applying changes", server.enforcementIdx)
		helmIstio := helm.NewPlugin(server.cfg.Helm)
		pluginRegistry = plugin.NewRegistry(
			[]plugin.DeployPlugin{helmIstio},
			[]plugin.PostProcessPlugin{helmIstio},
		)
	}

	eventLog = event.NewLog(fmt.Sprintf("enforce-%d-apply", server.enforcementIdx), true)
	applier := apply.NewEngineApply(desiredPolicy, desiredState, actualState, server.store.GetActualStateUpdater(), server.externalData, pluginRegistry, stateDiff.Actions, eventLog, server.store.GetRevisionProgressUpdater(nextRevision))
	_, err = applier.Apply()

	// todo save eventlog

	if err != nil {
		return fmt.Errorf("error while applying new revision: %s", err)
	}
	log.Infof("(enforce-%d) New revision %d successfully applied, %d component instances", server.enforcementIdx, nextRevision.GetGeneration(), len(desiredState.GetComponentProcessingOrder()))

	return nil
}
