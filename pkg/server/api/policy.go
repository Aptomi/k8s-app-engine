package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/server/api/reqresp"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (a *api) handlePolicyShow(r *http.Request, p httprouter.Params) reqresp.Response {
	gen := p.ByName("rev")

	if len(gen) == 0 {
		gen = string(object.LastGen)
	}

	policyData, err := a.store.GetPolicyData(object.ParseGeneration(gen))
	if err != nil {
		log.Panicf("error while getting requested policy: %s", err)
	}

	return policyData
}

func (a *api) handlePolicyUpdate(r *http.Request, p httprouter.Params) reqresp.Response {
	objects := a.read(r)

	changed, policyData, err := a.store.UpdatePolicy(objects)
	if err != nil {
		panic(fmt.Sprintf("Error while updating policy: %s", err))
	}

	if !changed {
		return nil
	}

	desiredPolicyGen := policyData.Generation
	desiredPolicy, _, err := a.store.GetPolicy(desiredPolicyGen)
	if err != nil {
		log.Panicf("Error while getting desiredPolicy: %s", err)
	}
	if desiredPolicy == nil {
		log.Panicf("Can't read policy right after updating it")
	}

	actualState, err := a.store.GetActualState()
	if err != nil {
		log.Panicf("Error while getting actual state: %s", err)
	}

	resolver := resolve.NewPolicyResolver(desiredPolicy, a.externalData)
	desiredState, eventLog, err := resolver.ResolveAllDependencies()
	if err != nil {
		log.Panicf("Cannot resolve desiredPolicy: %v %v %v", err, desiredState, actualState)
	}

	// todo save to log with clear prefix
	eventLog.Save(&event.HookStdout{})

	nextRevision, err := a.store.NextRevision(desiredPolicyGen)
	if err != nil {
		log.Panicf("Unable to get next revision: %s", err)
	}

	stateDiff := diff.NewPolicyResolutionDiff(desiredState, actualState, nextRevision.GetGeneration())

	return stateDiff.Actions
}
