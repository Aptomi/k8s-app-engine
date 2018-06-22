package server

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/runtime"
	log "github.com/sirupsen/logrus"
)

func (server *Server) actualStateUpdateLoop() error {
	for {
		err := server.actualStateUpdate()
		if err != nil {
			log.Errorf("error while updating actual state: %s", err)
		}

		// sleep for a specified time or wait until policy has changed, whichever comes first
		timer := time.NewTimer(server.cfg.Updater.Interval)
		select {
		case <-server.runActualStateUpdate:
			break // nolint: megacheck
		case <-timer.C:
			break // nolint: megacheck
		}
		timer.Stop()
	}
}

func (server *Server) actualStateUpdate() error {
	server.actualStateUpdateIdx++

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic while updating actual state: %s", err)
		}
	}()

	// Get desired policy
	desiredPolicy, _, err := server.store.GetPolicy(runtime.LastGen)
	if err != nil {
		return fmt.Errorf("error while getting last policy: %s", err)
	}

	// if policy is not found, it means it somehow was not initialized correctly. let's return error
	if desiredPolicy == nil {
		return fmt.Errorf("last policy is nil, does not exist in the store")
	}

	// Get actual state
	actualState, err := server.store.GetActualState()
	if err != nil {
		return fmt.Errorf("error while getting actual state: %s", err)
	}

	// Make an event log
	eventLog := event.NewLog(log.DebugLevel, fmt.Sprintf("update-%d", server.actualStateUpdateIdx)).AddConsoleHook(server.cfg.GetLogLevel())

	// Load endpoints for all components
	refreshEndpoints(desiredPolicy, actualState, server.store.NewActualStateUpdater(actualState), server.updaterPluginRegistryFactory(), eventLog, server.cfg.Updater.MaxConcurrentActions, server.cfg.Updater.Noop)

	log.Infof("(update-%d) Actual state updated", server.actualStateUpdateIdx)

	return nil
}

func refreshEndpoints(desiredPolicy *lang.Policy, actualState *resolve.PolicyResolution, actualStateUpdater actual.StateUpdater, plugins plugin.Registry, eventLog *event.Log, maxConcurrentActions int, noop bool) {
	context := action.NewContext(
		desiredPolicy,
		nil, // not needed for endpoints action
		actualStateUpdater,
		nil, // not needed for endpoints action
		plugins,
		eventLog,
	)

	// make sure we are converting panics into errors
	fn := action.WrapParallelWithLimit(maxConcurrentActions, func(act action.Interface) (errResult error) {
		defer func() {
			if err := recover(); err != nil {
				errResult = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
			}
		}()
		err := act.Apply(context)
		if err != nil {
			context.EventLog.NewEntry().Errorf("error while applying action '%s': %s", act, err)
		}
		return err
	})

	// generate the list of actions
	actions := []action.Interface{}
	for _, instance := range actualState.ComponentInstanceMap {
		if instance.IsCode && !instance.EndpointsUpToDate {
			var act action.Interface
			if !noop {
				act = component.NewEndpointsAction(instance.GetKey())
			} else {
				act = component.NewEndpointsAction(instance.GetKey())
			}
			actions = append(actions, act)
		}
	}

	// run actions
	var wg sync.WaitGroup
	for _, act := range actions {
		wg.Add(1)
		go func(act action.Interface) {
			defer wg.Done()

			// if an error or panic happened in the action, we don't have to do anything special, we will just retry it next time
			fn(act) // nolint: errcheck
		}(act)
	}

	// wait until all go routines are over
	wg.Wait()
}
