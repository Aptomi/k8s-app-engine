package slinga

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"time"
)

// ServiceUsageUserAction is a <ComponentKey, User> object. It holds data for attach/detach operations for user<->service
type ServiceUsageUserAction struct {
	ComponentKey string
	User         string
}

// ServiceUsageStateDiff represents a difference between two calculated usage states
type ServiceUsageStateDiff struct {
	// Pointers to previous and next states
	Prev *ServiceUsageState
	Next *ServiceUsageState

	// Actions that need to be taken
	ComponentInstantiate map[string]bool
	ComponentDestruct    map[string]bool
	ComponentUpdate      map[string]bool
	ComponentAttachUser  []ServiceUsageUserAction
	ComponentDetachUser  []ServiceUsageUserAction

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

func (usage ServiceUsageState) printSummaryLine(name string, cnt int) {
	fmt.Printf("  %s: %d\n", name, cnt)
}

// On a service level -- see which keys appear and disappear
func (diff *ServiceUsageStateDiff) printDifferenceOnServicesLevel(verbose bool) {

	fmt.Println("[Services]")

	// High-level service resolutions in prev (user -> serviceName -> serviceKey -> count)
	pMap := make(map[string]map[string]map[string]int)
	if diff.Prev.Dependencies != nil {
		for _, deps := range diff.Prev.Dependencies.Dependencies {
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
	for _, deps := range diff.Next.Dependencies.Dependencies {
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
	for k := range diff.Prev.getResolvedUsage().ComponentInstanceMap {
		allKeys[k] = true
	}
	for k := range diff.Next.getResolvedUsage().ComponentInstanceMap {
		allKeys[k] = true
	}
	// Go over all the keys and see which one appear and which one disappear
	for k := range allKeys {
		uPrev := diff.Prev.getResolvedUsage().ComponentInstanceMap[k]
		uNext := diff.Next.getResolvedUsage().ComponentInstanceMap[k]

		var userIdsPrev []string
		if uPrev != nil {
			userIdsPrev = uPrev.UserIds
		}

		var userIdsNext []string
		if uNext != nil {
			userIdsNext = uNext.UserIds
		}

		// see if a component needs to be instantiated
		if userIdsPrev == nil && userIdsNext != nil {
			diff.ComponentInstantiate[k] = true
		}

		// see if a component needs to be destructed
		if userIdsPrev != nil && userIdsNext == nil {
			diff.ComponentDestruct[k] = true
		}

		// see if a component needs to be updated
		if userIdsPrev != nil && userIdsNext != nil {
			sameParams := uPrev.CalculatedCodeParams.deepEqual(uNext.CalculatedCodeParams)
			if !sameParams {
				diff.ComponentUpdate[k] = true
			}
		}

		// see what needs to happen to users
		uPrevIdsMap := toMap(userIdsPrev)
		uNextIdsMap := toMap(userIdsNext)

		// see if a user needs to be detached from a component
		for u := range uPrevIdsMap {
			if !uNextIdsMap[u] {
				diff.ComponentDetachUser = append(diff.ComponentDetachUser, ServiceUsageUserAction{ComponentKey: k, User: u})
			}
		}

		// see if a user needs to be attached to a component
		for u := range uNextIdsMap {
			if !uPrevIdsMap[u] {
				diff.ComponentAttachUser = append(diff.ComponentAttachUser, ServiceUsageUserAction{ComponentKey: k, User: u})
			}
		}
	}
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
		for key := range diff.Next.getResolvedUsage().ComponentInstanceMap {
			diff.ComponentInstantiate[key] = true
		}
	}
}

// Apply method applies all changes via executors and saves usage state in Aptomi DB
func (diff ServiceUsageStateDiff) Apply(noop bool) {
	if !noop {
		// add progress bar into CLI
		dLen := diff.getDifferenceLen()
		if dLen > 0 {
			fmt.Println("[Applying changes]")
			diff.progress = uiprogress.New()
			diff.progress.RefreshInterval = time.Second
			diff.progress.Start()
			diff.progressBar = diff.progress.AddBar(dLen)
			diff.progressBar.PrependFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("  [%d/%d]", b.Current(), b.Total)
			})
			diff.progressBar.AppendCompleted()
			diff.progressBar.AppendFunc(func(b *uiprogress.Bar) string {
				return fmt.Sprintf("  Time: %s", strutil.PrettyTime(time.Since(b.TimeStarted)))
			})
		}

		err := diff.processDestructions()
		if err != nil {
			debug.WithFields(log.Fields{
				"error": err,
			}).Fatal("Error while destructing components")
		}
		err = diff.processUpdates()
		if err != nil {
			debug.WithFields(log.Fields{
				"error": err,
			}).Fatal("Error while updating components")
		}
		err = diff.processInstantiations()
		if err != nil {
			debug.WithFields(log.Fields{
				"error": err,
			}).Fatal("Error while instantiating components")
		}

		// Don't forget to stop the progress bar and print its final state
		if diff.progress != nil {
			diff.progress.Stop()
		}
	}

	// save new state
	diff.Next.SaveServiceUsageState(noop)
}

func (diff ServiceUsageStateDiff) processInstantiations() error {
	// Process instantiations in the right order
	for _, key := range diff.Next.getResolvedUsage().ComponentProcessingOrder {
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
					codeExecutor, err := component.Code.GetCodeExecutor(key, component.Code.Metadata, diff.Next.getResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, diff.Next.Policy.Clusters)
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
	for _, key := range diff.Next.getResolvedUsage().ComponentProcessingOrder {
		// Does it need to be updated?
		if _, ok := diff.ComponentUpdate[key]; ok {
			// Increment progress bar
			if diff.progressBar != nil {
				diff.progressBar.Incr()
			}

			serviceName, _ /*contextName*/, _ /*allocationName*/, componentName := ParseServiceUsageKey(key)
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
					codeExecutor, err := component.Code.GetCodeExecutor(key, component.Code.Metadata, diff.Next.getResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, diff.Next.Policy.Clusters)
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
	for _, key := range diff.Prev.getResolvedUsage().ComponentProcessingOrder {
		// Does it need to be destructed?
		if _, ok := diff.ComponentDestruct[key]; ok {
			// Increment progress bar
			if diff.progressBar != nil {
				diff.progressBar.Incr()
			}

			serviceName, _ /*contextName*/, _ /*allocationName*/, componentName := ParseServiceUsageKey(key)
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
					codeExecutor, err := component.Code.GetCodeExecutor(key, component.Code.Metadata, diff.Prev.getResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, diff.Prev.Policy.Clusters)
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
