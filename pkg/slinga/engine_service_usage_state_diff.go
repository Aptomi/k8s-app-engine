package slinga

import (
	"fmt"
	. "github.com/Frostman/aptomi/pkg/slinga/db"
	. "github.com/Frostman/aptomi/pkg/slinga/log"
	log "github.com/Sirupsen/logrus"
	"github.com/gosuri/uiprogress"
	"time"
)

// ServiceUsageDependencyAction is a <ComponentKey, DependencyID> object. It holds data for attach/detach operations for component instance <-> dependency
type ServiceUsageDependencyAction struct {
	ComponentKey string
	DependencyID string
}

// ServiceUsageStateDiff represents a difference between two calculated usage states
type ServiceUsageStateDiff struct {
	// Pointers to previous and next states
	Prev *ServiceUsageState
	Next *ServiceUsageState

	// Actions that need to be taken
	ComponentInstantiate      map[string]bool
	ComponentDestruct         map[string]bool
	ComponentUpdate           map[string]bool
	ComponentAttachDependency []ServiceUsageDependencyAction
	ComponentDetachDependency []ServiceUsageDependencyAction

	// Progress bar for CLI
	progress    *uiprogress.Progress
	progressBar *uiprogress.Bar
}

// CalculateDifference calculates difference between two given usage states
func (state *ServiceUsageState) CalculateDifference(prev *ServiceUsageState) *ServiceUsageStateDiff {
	// resulting difference
	result := &ServiceUsageStateDiff{
		Prev:                 prev,
		Next:                 state,
		ComponentInstantiate: make(map[string]bool),
		ComponentDestruct:    make(map[string]bool),
		ComponentUpdate:      make(map[string]bool)}

	result.calculateDifferenceOnComponentLevel()

	return result
}

// ServiceUsageStateSummary returns integer counts for all policy objects
type ServiceUsageStateSummary struct {
	Services     int
	Contexts     int
	Clusters     int
	Rules        int
	UserText     string
	Dependencies int
}

// compare everything, but not users. users are external to us
func (summary ServiceUsageStateSummary) equal(that ServiceUsageStateSummary) bool {
	if summary.Services != that.Services {
		return false
	}
	if summary.Contexts != that.Contexts {
		return false
	}
	if summary.Clusters != that.Clusters {
		return false
	}
	if summary.Rules != that.Rules {
		return false
	}
	if summary.Dependencies != that.Dependencies {
		return false
	}
	return true
}

// GetSummary returns summary object for the policy
func (state ServiceUsageState) GetSummary() ServiceUsageStateSummary {
	if state.Policy == nil {
		return ServiceUsageStateSummary{}
	}
	return ServiceUsageStateSummary{
		state.Policy.CountServices(),
		state.Policy.CountContexts(),
		state.Policy.CountClusters(),
		state.Policy.Rules.Count(),
		state.userLoader.Summary(),
		state.Dependencies.Count(),
	}
}

// ProcessSuccessfulExecution increments revision and saves results of the current run when policy processing executed successfully
func (diff *ServiceUsageStateDiff) ProcessSuccessfulExecution(revision AptomiRevision, newrevision bool, noop bool) {
	fmt.Println("[Revision]")
	if newrevision || (!noop && diff.ShouldGenerateNewRevision()) {
		// Increment a revision
		newRevision := revision.Increment()

		// Save results of the current run
		newRevision.SaveCurrentRun()

		// Save updated revision number
		newRevision.SaveAsLastRevision()

		// Print revision numbers
		fmt.Printf("  Previous: %s\n", revision.String())
		fmt.Printf("  Current: %s\n", newRevision.String())
	} else {
		fmt.Printf("  Current: %s (no changes made)\n", revision.String())
	}
}

// On a service level -- see which keys appear and disappear
func (diff *ServiceUsageStateDiff) writeDifferenceOnServicesLevel(log *log.Logger) {

	log.Println("[Services]")

	// High-level service resolutions in prev (user -> serviceName -> serviceKey -> count)
	pMap := make(map[string]map[string]map[string]int)
	if diff.Prev.Dependencies != nil {
		for _, deps := range diff.Prev.Dependencies.DependenciesByService {
			for _, d := range deps {
				// Make sure to check for the case when service hasn't been resolved (no matching context/allocation found)
				if d.Resolved {
					if pMap[d.UserID] == nil {
						pMap[d.UserID] = make(map[string]map[string]int)
					}
					if pMap[d.UserID][d.Service] == nil {
						pMap[d.UserID][d.Service] = make(map[string]int)
					}
					pMap[d.UserID][d.Service][d.ServiceKey]++
				}
			}
		}
	}

	// High-level service resolutions in next (user -> serviceName -> serviceKey -> count)
	cMap := make(map[string]map[string]map[string]int)
	for _, deps := range diff.Next.Dependencies.DependenciesByService {
		for _, d := range deps {
			// Make sure to check for the case when service hasn't been resolved (no matching context/allocation found)
			if d.Resolved {
				if cMap[d.UserID] == nil {
					cMap[d.UserID] = make(map[string]map[string]int)
				}
				if cMap[d.UserID][d.Service] == nil {
					cMap[d.UserID][d.Service] = make(map[string]int)
				}
				cMap[d.UserID][d.Service][d.ServiceKey]++
			}
		}
	}

	// map of all user keys
	userKeyMap := make(map[string]bool)
	for userID := range pMap {
		userKeyMap[userID] = true
	}
	for userID := range cMap {
		userKeyMap[userID] = true
	}

	// Printable description
	textMap := make(map[string][]string)

	// Go over all users
	for userID := range userKeyMap {
		sPrev := pMap[userID]
		sNext := cMap[userID]

		// merge all the service names
		serviceNameMap := make(map[string]bool)
		for serviceName := range sPrev {
			serviceNameMap[serviceName] = true
		}
		for serviceName := range sNext {
			serviceNameMap[serviceName] = true
		}

		// For every service
		for serviceName := range serviceNameMap {
			// Figure out how many service keys got added vs got removed
			prevMap := sPrev[serviceName]
			nextMap := sNext[serviceName]

			// Merge all the service keys
			serviceKeyMap := make(map[string]bool)
			for serviceKey := range prevMap {
				serviceKeyMap[serviceKey] = true
			}
			for serviceKey := range nextMap {
				serviceKeyMap[serviceKey] = true
			}

			// For every service key
			for serviceKey := range serviceKeyMap {
				prevCnt := prevMap[serviceKey]
				nextCnt := nextMap[serviceKey]

				for i := 0; i < nextCnt-prevCnt; i++ {
					textMap[userID] = append(textMap[userID], fmt.Sprintf("[+] %s (%s)", serviceName, serviceKey))
				}

				for i := 0; i < prevCnt-nextCnt; i++ {
					textMap[userID] = append(textMap[userID], fmt.Sprintf("[-] %s (%s)", serviceName, serviceKey))
				}
			}

		}
	}

	// Print
	printed := false
	for userID, sKeys := range textMap {
		user := diff.Next.userLoader.LoadUserByID(userID)
		log.Printf("  %s (ID=%s)", user.Name, user.ID)
		for _, s := range sKeys {
			log.Printf("    %s", s)
		}
		printed = true
	}

	if !printed {
		log.Println("  [*] No changes")
	}
}

// On a component level -- see which allocation keys appear and disappear
func (diff *ServiceUsageStateDiff) calculateDifferenceOnComponentLevel() {
	// map of all instances
	allKeys := make(map[string]bool)

	// merge all the keys
	for k := range diff.Prev.GetResolvedData().ComponentInstanceMap {
		allKeys[k] = true
	}
	for k := range diff.Next.GetResolvedData().ComponentInstanceMap {
		allKeys[k] = true
	}
	// Go over all the keys and see which one appear and which one disappear
	for k := range allKeys {
		uPrev := diff.Prev.GetResolvedData().ComponentInstanceMap[k]
		uNext := diff.Next.GetResolvedData().ComponentInstanceMap[k]

		var depIdsPrev map[string]bool
		if uPrev != nil {
			depIdsPrev = uPrev.DependencyIds
		}

		var depIdsNext map[string]bool
		if uNext != nil {
			depIdsNext = uNext.DependencyIds
		}

		// see if a component needs to be instantiated
		if len(depIdsPrev) <= 0 && len(depIdsNext) > 0 {
			diff.ComponentInstantiate[k] = true
			diff.updateTimes(k, time.Now(), time.Now())
		}

		// see if a component needs to be destructed
		if len(depIdsPrev) > 0 && len(depIdsNext) <= 0 {
			diff.ComponentDestruct[k] = true
			diff.updateTimes(k, uPrev.CreatedOn, time.Now())
		}

		// see if a component needs to be updated
		if len(depIdsPrev) > 0 && len(depIdsNext) > 0 {
			sameParams := uPrev.CalculatedCodeParams.DeepEqual(uNext.CalculatedCodeParams)
			if !sameParams {
				diff.ComponentUpdate[k] = true
				diff.updateTimes(k, uPrev.CreatedOn, time.Now())
			} else {
				diff.updateTimes(k, uPrev.CreatedOn, uPrev.UpdatedOn)
			}
		}

		// see if a user needs to be detached from a component
		for dependencyID := range depIdsPrev {
			if !depIdsNext[dependencyID] {
				diff.ComponentDetachDependency = append(diff.ComponentDetachDependency, ServiceUsageDependencyAction{ComponentKey: k, DependencyID: dependencyID})
			}
		}

		// see if a user needs to be attached to a component
		for dependencyID := range depIdsNext {
			if !depIdsPrev[dependencyID] {
				diff.ComponentAttachDependency = append(diff.ComponentAttachDependency, ServiceUsageDependencyAction{ComponentKey: k, DependencyID: dependencyID})
			}
		}
	}
}

// updated timestamps for component (and root service, if/as needed)
func (diff *ServiceUsageStateDiff) updateTimes(k string, createdOn time.Time, updatedOn time.Time) {
	// update for a given node
	instance := diff.Next.GetResolvedData().ComponentInstanceMap[k]
	if instance == nil {
		// likely this component has been deleted as a part of the diff
		return
	}
	instance.updateTimes(createdOn, updatedOn)

	// if it's a component instance, then update for its parent service instance as well
	serviceName, contextName, allocationName, componentName := ParseServiceUsageKey(k)
	if componentName != ComponentRootName {
		serviceKey := createServiceUsageKeyFromStr(serviceName, contextName, allocationName, ComponentRootName)
		diff.updateTimes(serviceKey, createdOn, updatedOn)
	}
}

// ShouldGenerateNewRevision became one of the key methods.
// If it returns false, then new run and new revision will not be generated
func (diff *ServiceUsageStateDiff) ShouldGenerateNewRevision() bool {
	if !diff.Next.GetSummary().equal(diff.Prev.GetSummary()) {
		return true
	}
	if len(diff.ComponentInstantiate) > 0 {
		return true
	}
	if len(diff.ComponentUpdate) > 0 {
		return true
	}
	if len(diff.ComponentDestruct) > 0 {
		return true
	}
	if len(diff.ComponentAttachDependency) > 0 {
		return true
	}
	if len(diff.ComponentDetachDependency) > 0 {
		return true
	}

	// TODO: this key method doesn't take into account presence of Istio rules
	return false
}

// On a component level -- see which keys appear and disappear
func (diff *ServiceUsageStateDiff) getDifferenceLen() int {
	return len(diff.ComponentInstantiate) + len(diff.ComponentDestruct) + len(diff.ComponentUpdate)
}

// On a component level -- see which keys appear and disappear
func (diff *ServiceUsageStateDiff) writeDifferenceOnComponentLevel(verbose bool, log *log.Logger) {
	log.Println("[Components]")

	// Print
	printed := diff.getDifferenceLen() > 0

	if printed {
		log.Printf("  New instances:     %d", len(diff.ComponentInstantiate))
		log.Printf("  Deleted instances: %d", len(diff.ComponentDestruct))
		log.Printf("  Updated instances: %d", len(diff.ComponentUpdate))

		log.Println("[Component Instances]")
		if verbose {
			if len(diff.ComponentInstantiate) > 0 {
				log.Debug("  New:")
				for k := range diff.ComponentInstantiate {
					log.Printf("    [+] %s", k)
				}
			}

			if len(diff.ComponentDestruct) > 0 {
				log.Debug("  Deleted:")
				for k := range diff.ComponentDestruct {
					log.Printf("    [-] %s", k)
				}
			}

			if len(diff.ComponentUpdate) > 0 {
				log.Debug("  Updated:")
				for k := range diff.ComponentUpdate {
					log.Printf("    [*] %s", k)
				}
			}
		} else {
			log.Println("  Use --verbose to see the list")
		}
	}

	if !printed {
		log.Println("  [*] No changes")
	}

}

// Returns summary line with two object counts
func getSummaryLine(name string, cntPrev int, cntNext int) string {
	if cntPrev != cntNext {
		delta := ""
		if cntNext-cntPrev != 0 {
			delta = fmt.Sprintf(" [%+d]", cntNext-cntPrev)
		}
		return fmt.Sprintf("  %s: %d%s", name, cntNext, delta)
	}
	return fmt.Sprintf("  %s: %d", name, cntPrev)
}

// Returns summary line (as we comparing text-based data, such as summary for external users)
func getSummaryLineAsText(name string, prevStr string, nextStr string) string {
	if nextStr != prevStr && len(prevStr) > 0 {
		return fmt.Sprintf("  %s: %s (was %s)", name, nextStr, prevStr)
	}
	return fmt.Sprintf("  %s: %s", name, nextStr)
}

// PrintSummary prints policy object counts to the screen
func (diff ServiceUsageStateDiff) writeSummary(log *log.Logger) {
	prev := diff.Prev.GetSummary()
	next := diff.Next.GetSummary()
	log.Println("[Policy]")
	log.Println(getSummaryLine("Services", prev.Services, next.Services))
	log.Println(getSummaryLine("Contexts", prev.Contexts, next.Contexts))
	log.Println(getSummaryLine("Dependencies", prev.Dependencies, next.Dependencies))
	log.Println(getSummaryLine("Clusters", prev.Clusters, next.Clusters))
	log.Println(getSummaryLine("Rules", prev.Rules, next.Rules))
	log.Println(getSummaryLineAsText("Users", prev.UserText, next.UserText))
}

// StoreDiffAsText method prints changes onto the screen (i.e. delta - what got added/removed)
func (diff ServiceUsageStateDiff) StoreDiffAsText(verbose bool) {
	memLog := NewPlainMemoryLogger(verbose)
	diff.writeSummary(memLog.GetLogger())
	diff.writeDifferenceOnServicesLevel(memLog.GetLogger())
	diff.writeDifferenceOnComponentLevel(verbose, memLog.GetLogger())
	diff.Next.DiffAsText = memLog.GetBuffer().String()
}

// AlterDifference basically marks all objects in diff for recreation/update if full update is requested. This is useful to re-create missing objects if they got deleted externally after deployment
func (diff *ServiceUsageStateDiff) AlterDifference(full bool) {
	// If we are requesting full policy processing, then we will need to re-create all objects
	if full {
		for k, v := range diff.Next.GetResolvedData().ComponentInstanceMap {
			diff.ComponentInstantiate[k] = true
			diff.updateTimes(k, v.CreatedOn, time.Now())
		}
	}
}

// Apply method applies all changes via executors, saves usage state in Aptomi DB
func (diff ServiceUsageStateDiff) Apply(noop bool) {
	if !noop {
		// add progress bar into CLI
		dLen := diff.getDifferenceLen()
		if dLen > 0 {
			fmt.Println("[Applying changes]")
			diff.progress = NewProgress()
			diff.progressBar = AddProgressBar(diff.progress, dLen)
		}

		err := diff.processDestructions()
		if err != nil {
			Debug.WithFields(log.Fields{
				"error": err,
			}).Panic("Error while destructing components")
		}
		err = diff.processUpdates()
		if err != nil {
			Debug.WithFields(log.Fields{
				"error": err,
			}).Panic("Error while updating components")
		}
		err = diff.processInstantiations()
		if err != nil {
			Debug.WithFields(log.Fields{
				"error": err,
			}).Panic("Error while instantiating components")
		}

		// Don't forget to stop the progress bar and print its final state
		if diff.progress != nil {
			diff.progress.Stop()
		}

		// Apply changes in Istio Ingress rules
		diff.Next.ProcessIstioIngress(noop)
	}

	// save new state in the last run directory
	diff.Next.SaveServiceUsageState()
}

func (diff ServiceUsageStateDiff) processInstantiations() error {
	// Process instantiations in the right order
	for _, key := range diff.Next.GetResolvedData().ComponentProcessingOrder {
		// Does it need to be instantiated?
		if _, ok := diff.ComponentInstantiate[key]; ok {
			// Increment progress bar
			if diff.progressBar != nil {
				diff.progressBar.Incr()
			}

			serviceName, _, _, componentName := ParseServiceUsageKey(key)
			component := diff.Next.Policy.Services[serviceName].GetComponentsMap()[componentName]

			if component == nil {
				Debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    serviceName,
				}).Info("Instantiating service")

				// TODO: add processing code
			} else {
				Debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Instantiating component")

				if component.Code != nil {
					codeExecutor, err := GetCodeExecutor(component.Code, key, diff.Next.GetResolvedData().ComponentInstanceMap[key].CalculatedCodeParams, diff.Next.Policy.Clusters)
					if err != nil {
						return err
					}

					err = codeExecutor.Install()
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (diff ServiceUsageStateDiff) processUpdates() error {
	// Process updates in the right order
	for _, key := range diff.Next.GetResolvedData().ComponentProcessingOrder {
		// Does it need to be updated?
		if _, ok := diff.ComponentUpdate[key]; ok {
			// Increment progress bar
			if diff.progressBar != nil {
				diff.progressBar.Incr()
			}

			serviceName, _ /*contextName*/, _ /*allocationName*/, componentName := ParseServiceUsageKey(key)
			component := diff.Prev.Policy.Services[serviceName].GetComponentsMap()[componentName]
			if component == nil {
				Debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    serviceName,
				}).Info("Updating service")

				// TODO: add processing code
			} else {
				Debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Updating component")

				if component.Code != nil {
					codeExecutor, err := GetCodeExecutor(component.Code, key, diff.Next.GetResolvedData().ComponentInstanceMap[key].CalculatedCodeParams, diff.Next.Policy.Clusters)
					if err != nil {
						return err
					}
					err = codeExecutor.Update()
					if err != nil {
						return err
					}

				}
			}
		}
	}
	return nil
}

func (diff ServiceUsageStateDiff) processDestructions() error {
	// Process destructions in the right order
	for _, key := range diff.Prev.GetResolvedData().ComponentProcessingOrder {
		// Does it need to be destructed?
		if _, ok := diff.ComponentDestruct[key]; ok {
			// Increment progress bar
			if diff.progressBar != nil {
				diff.progressBar.Incr()
			}

			serviceName, _ /*contextName*/, _ /*allocationName*/, componentName := ParseServiceUsageKey(key)
			component := diff.Prev.Policy.Services[serviceName].GetComponentsMap()[componentName]
			if component == nil {
				Debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    serviceName,
				}).Info("Destructing service")

				// TODO: add processing code
			} else {
				Debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Destructing component")

				if component.Code != nil {
					codeExecutor, err := GetCodeExecutor(component.Code, key, diff.Prev.GetResolvedData().ComponentInstanceMap[key].CalculatedCodeParams, diff.Prev.Policy.Clusters)
					if err != nil {
						return err
					}
					err = codeExecutor.Destroy()
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
