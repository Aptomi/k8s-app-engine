package slinga

import (
	"fmt"
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
func (usage *ServiceUsageState) CalculateDifference(prev *ServiceUsageState) *ServiceUsageStateDiff {
	// resulting difference
	result := &ServiceUsageStateDiff{
		Prev:                 prev,
		Next:                 usage,
		ComponentInstantiate: make(map[string]bool),
		ComponentDestruct:    make(map[string]bool),
		ComponentUpdate:      make(map[string]bool)}

	result.calculateDifferenceOnComponentLevel()

	return result
}

// PrintSummary prints policy object counts to the screen
func (usage ServiceUsageState) PrintSummary() {
	fmt.Println("[Policy]")
	usage.printSummaryLine("Services", usage.Policy.countServices())
	usage.printSummaryLine("Contexts", usage.Policy.countContexts())
	usage.printSummaryLine("Clusters", usage.Policy.countClusters())
	usage.printSummaryLine("Rules", usage.Policy.Rules.count())
	usage.printSummaryLine("Users", usage.users.count())
	usage.printSummaryLine("Dependencies", usage.Dependencies.count())
}

// ProcessSuccessfulExecution increments revision and saves results of the current run when policy processing executed successfully
func (diff *ServiceUsageStateDiff) ProcessSuccessfulExecution(revision AptomiRevision, forceSave bool, noop bool) {
	fmt.Println("[Revision]")
	if forceSave || (!noop && diff.hasChanges()) {
		// Increment a revision
		newRevision := revision.increment()

		// Save results of the current run
		newRevision.saveCurrentRun()

		// Save updated revision number
		newRevision.saveAsLastRevision()

		// Print revision numbers
		fmt.Printf("  Previous: %s\n", revision.String())
		fmt.Printf("  Current: %s\n", newRevision.String())
	} else {
		fmt.Printf("  Current: %s (no changes made)\n", revision.String())
	}
}

func (usage ServiceUsageState) printSummaryLine(name string, cnt int) {
	fmt.Printf("  %s: %d\n", name, cnt)
}

// On a service level -- see which keys appear and disappear
func (diff *ServiceUsageStateDiff) printDifferenceOnServicesLevel(verbose bool) {

	fmt.Println("[Services]")

	// High-level service resolutions in prev (user -> serviceName -> serviceKey -> count)
	pMap := make(map[string]map[string]map[string]int)
	if diff.Prev.Dependencies != nil {
		for _, deps := range diff.Prev.Dependencies.DependenciesByService {
			for _, d := range deps {
				// Make sure to check for the case when service hasn't been resolved (no matching context/allocation found)
				if len(d.ResolvesTo) > 0 {
					if pMap[d.UserID] == nil {
						pMap[d.UserID] = make(map[string]map[string]int)
					}
					if pMap[d.UserID][d.Service] == nil {
						pMap[d.UserID][d.Service] = make(map[string]int)
					}
					pMap[d.UserID][d.Service][d.ResolvesTo]++
				}
			}
		}
	}

	// High-level service resolutions in next (user -> serviceName -> serviceKey -> count)
	cMap := make(map[string]map[string]map[string]int)
	for _, deps := range diff.Next.Dependencies.DependenciesByService {
		for _, d := range deps {
			// Make sure to check for the case when service hasn't been resolved (no matching context/allocation found)
			if len(d.ResolvesTo) > 0 {
				if cMap[d.UserID] == nil {
					cMap[d.UserID] = make(map[string]map[string]int)
				}
				if cMap[d.UserID][d.Service] == nil {
					cMap[d.UserID][d.Service] = make(map[string]int)
				}
				cMap[d.UserID][d.Service][d.ResolvesTo]++
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
		user := LoadUserByIDFromDir(GetAptomiBaseDir(), userID)
		fmt.Printf("  %s (ID=%s)\n", user.Name, user.ID)
		for _, s := range sKeys {
			fmt.Printf("    %s\n", s)
		}
		printed = true
	}

	if !printed {
		fmt.Println("  [*] No changes")
	}
}

// On a component level -- see which allocation keys appear and disappear
func (diff *ServiceUsageStateDiff) calculateDifferenceOnComponentLevel() {
	// map of all instances
	allKeys := make(map[string]bool)

	// merge all the keys
	for k := range diff.Prev.GetResolvedUsage().ComponentInstanceMap {
		allKeys[k] = true
	}
	for k := range diff.Next.GetResolvedUsage().ComponentInstanceMap {
		allKeys[k] = true
	}
	// Go over all the keys and see which one appear and which one disappear
	for k := range allKeys {
		uPrev := diff.Prev.GetResolvedUsage().ComponentInstanceMap[k]
		uNext := diff.Next.GetResolvedUsage().ComponentInstanceMap[k]

		var depIdsPrev []string
		if uPrev != nil {
			depIdsPrev = uPrev.DependencyIds
		}

		var depIdsNext []string
		if uNext != nil {
			depIdsNext = uNext.DependencyIds
		}

		// see if a component needs to be instantiated
		if depIdsPrev == nil && depIdsNext != nil {
			diff.ComponentInstantiate[k] = true
			diff.updateTimes(k, time.Now(), time.Now())
		}

		// see if a component needs to be destructed
		if depIdsPrev != nil && depIdsNext == nil {
			diff.ComponentDestruct[k] = true
			diff.updateTimes(k, uPrev.CreatedOn, time.Now())
		}

		// see if a component needs to be updated
		if depIdsPrev != nil && depIdsNext != nil {
			sameParams := uPrev.CalculatedCodeParams.deepEqual(uNext.CalculatedCodeParams)
			if !sameParams {
				diff.ComponentUpdate[k] = true
				diff.updateTimes(k, uPrev.CreatedOn, time.Now())
			} else {
				diff.updateTimes(k, uPrev.CreatedOn, uPrev.UpdatedOn)
			}
		}

		// see what needs to happen to users
		depPrevIdsMap := toMap(depIdsPrev)
		depNextIdsMap := toMap(depIdsNext)

		// see if a user needs to be detached from a component
		for dependencyId := range depPrevIdsMap {
			if !depNextIdsMap[dependencyId] {
				diff.ComponentDetachDependency = append(diff.ComponentDetachDependency, ServiceUsageDependencyAction{ComponentKey: k, DependencyID: dependencyId})
			}
		}

		// see if a user needs to be attached to a component
		for dependencyId := range depNextIdsMap {
			if !depPrevIdsMap[dependencyId] {
				diff.ComponentAttachDependency = append(diff.ComponentAttachDependency, ServiceUsageDependencyAction{ComponentKey: k, DependencyID: dependencyId})
			}
		}
	}
}

// updated timestamps for component (and root service, if/as needed)
func (diff *ServiceUsageStateDiff) updateTimes(k string, createdOn time.Time, updatedOn time.Time) {
	// update for a given node
	instance := diff.Next.GetResolvedUsage().ComponentInstanceMap[k]
	if createdOn.After(instance.CreatedOn) {
		instance.CreatedOn = createdOn
	}
	if updatedOn.After(instance.UpdatedOn) {
		instance.UpdatedOn = updatedOn
	}

	// if it's a component instance, then update for its parent service instance as well
	serviceName, contextName , allocationName, componentName := ParseServiceUsageKey(k)
	if componentName != ComponentRootName {
		kService := createServiceUsageKeyFromStr(serviceName, contextName , allocationName, ComponentRootName)
		diff.updateTimes(kService, createdOn, updatedOn)
	}
}

// This method became one of the key methods.
// If it returns false, then new run and new revision will not be generated
func (diff *ServiceUsageStateDiff) hasChanges() bool {
	// TODO: this key method doesn't take into account presence of Istio rules
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
	return false
}

// On a component level -- see which keys appear and disappear
func (diff *ServiceUsageStateDiff) getDifferenceLen() int {
	return len(diff.ComponentInstantiate) + len(diff.ComponentDestruct) + len(diff.ComponentUpdate)
}

// On a component level -- see which keys appear and disappear
func (diff *ServiceUsageStateDiff) printDifferenceOnComponentLevel(verbose bool) {
	fmt.Println("[Components]")

	// Print
	printed := diff.getDifferenceLen() > 0

	if printed {
		fmt.Printf("  New instances:     %d\n", len(diff.ComponentInstantiate))
		fmt.Printf("  Deleted instances: %d\n", len(diff.ComponentDestruct))
		fmt.Printf("  Updated instances: %d\n", len(diff.ComponentUpdate))

		fmt.Println("[Component Instances]")
		if verbose {
			if len(diff.ComponentInstantiate) > 0 {
				fmt.Println("  New:")
				for k := range diff.ComponentInstantiate {
					fmt.Printf("    [+] %s\n", k)
				}
			}

			if len(diff.ComponentDestruct) > 0 {
				fmt.Println("  Deleted:")
				for k := range diff.ComponentDestruct {
					fmt.Printf("    [-] %s\n", k)
				}
			}

			if len(diff.ComponentUpdate) > 0 {
				fmt.Println("  Updated:")
				for k := range diff.ComponentUpdate {
					fmt.Printf("    [*] %s\n", k)
				}
			}
		} else {
			fmt.Println("  Use --verbose to see the list")
		}
	}

	if !printed {
		fmt.Println("  [*] No changes")
	}

}

func toMap(p []string) map[string]bool {
	result := make(map[string]bool)
	for _, s := range p {
		result[s] = true
	}
	return result
}

// Print method prints changes onto the screen (i.e. delta - what got added/removed)
func (diff ServiceUsageStateDiff) Print(verbose bool) {
	diff.printDifferenceOnServicesLevel(verbose)
	diff.printDifferenceOnComponentLevel(verbose)
}

// AlterDifference basically marks all objects in diff for recreation/update if full update is requested. This is useful to re-create missing objects if they got deleted externally after deployment
func (diff *ServiceUsageStateDiff) AlterDifference(full bool) {
	// If we are requesting full policy processing, then we will need to re-create all objects
	if full {
		for k, v := range diff.Next.GetResolvedUsage().ComponentInstanceMap {
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
			debug.WithFields(log.Fields{
				"error": err,
			}).Panic("Error while destructing components")
		}
		err = diff.processUpdates()
		if err != nil {
			debug.WithFields(log.Fields{
				"error": err,
			}).Panic("Error while updating components")
		}
		err = diff.processInstantiations()
		if err != nil {
			debug.WithFields(log.Fields{
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
	for _, key := range diff.Next.GetResolvedUsage().ComponentProcessingOrder {
		// Does it need to be instantiated?
		if _, ok := diff.ComponentInstantiate[key]; ok {
			// Increment progress bar
			if diff.progressBar != nil {
				diff.progressBar.Incr()
			}

			serviceName, _, _, componentName := ParseServiceUsageKey(key)
			component := diff.Next.Policy.Services[serviceName].getComponentsMap()[componentName]

			if component == nil {
				debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    serviceName,
				}).Info("Instantiating service")

				// TODO: add processing code
			} else {
				debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Instantiating component")

				if component.Code != nil {
					codeExecutor, err := component.Code.GetCodeExecutor(key, diff.Next.GetResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, diff.Next.Policy.Clusters)
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
	for _, key := range diff.Next.GetResolvedUsage().ComponentProcessingOrder {
		// Does it need to be updated?
		if _, ok := diff.ComponentUpdate[key]; ok {
			// Increment progress bar
			if diff.progressBar != nil {
				diff.progressBar.Incr()
			}

			serviceName, _ /*contextName*/ , _ /*allocationName*/ , componentName := ParseServiceUsageKey(key)
			component := diff.Prev.Policy.Services[serviceName].getComponentsMap()[componentName]
			if component == nil {
				debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    serviceName,
				}).Info("Updating service")

				// TODO: add processing code
			} else {
				debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Updating component")

				if component.Code != nil {
					codeExecutor, err := component.Code.GetCodeExecutor(key, diff.Next.GetResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, diff.Next.Policy.Clusters)
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
	for _, key := range diff.Prev.GetResolvedUsage().ComponentProcessingOrder {
		// Does it need to be destructed?
		if _, ok := diff.ComponentDestruct[key]; ok {
			// Increment progress bar
			if diff.progressBar != nil {
				diff.progressBar.Incr()
			}

			serviceName, _ /*contextName*/ , _ /*allocationName*/ , componentName := ParseServiceUsageKey(key)
			component := diff.Prev.Policy.Services[serviceName].getComponentsMap()[componentName]
			if component == nil {
				debug.WithFields(log.Fields{
					"serviceKey": key,
					"service":    serviceName,
				}).Info("Destructing service")

				// TODO: add processing code
			} else {
				debug.WithFields(log.Fields{
					"componentKey": key,
					"component":    component.Name,
					"code":         component.Code,
				}).Info("Destructing component")

				if component.Code != nil {
					codeExecutor, err := component.Code.GetCodeExecutor(key, diff.Prev.GetResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, diff.Prev.Policy.Clusters)
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
