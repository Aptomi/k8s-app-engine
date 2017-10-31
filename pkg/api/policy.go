package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/object"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func (a *api) handlePolicyShow(r *http.Request, p httprouter.Params) Response {
	gen := p.ByName("rev")

	if len(gen) == 0 {
		gen = strconv.Itoa(int(object.LastGen))
	}

	policyData, err := a.store.GetPolicyData(object.ParseGeneration(gen))
	if err != nil {
		log.Panicf("error while getting requested policy: %s", err)
	}

	return policyData
}

func (a *api) handlePolicyUpdate(r *http.Request, p httprouter.Params) Response {
	objects := a.read(r)

	username := r.Header.Get("Username")
	// todo check empty username
	user := a.externalData.UserLoader.LoadUserByName(username)
	// todo check user == nil

	// Verify ACL for updated objects
	policy, _, err := a.store.GetPolicy(object.LastGen)
	if err != nil {
		log.Panicf("Error while loading current policy: %s", err)
	}
	for _, obj := range objects {
		errAdd := policy.View(user).ManageObject(obj)
		if errAdd != nil {
			log.Panicf("Error while adding updated object to policy: %s", errAdd)
		}
	}

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

	// todo save to log with clear prefix
	eventLog.Save(&event.HookConsole{})

	if err != nil {
		log.Panicf("Cannot resolve desiredPolicy: %v %v %v", err, desiredState, actualState)
	}

	nextRevision, err := a.store.NextRevision(desiredPolicyGen)
	if err != nil {
		log.Panicf("Unable to get next revision: %s", err)
	}

	stateDiff := diff.NewPolicyResolutionDiff(desiredState, actualState, nextRevision.GetGeneration())

	return stateDiff.Actions
}
