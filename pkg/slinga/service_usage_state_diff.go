package slinga

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
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
	ComponentAttachUser  []ServiceUsageUserAction
	ComponentDetachUser  []ServiceUsageUserAction

	// High-level generated text for the diff
}

// CalculateDifference calculates difference between two given usage states
func (next *ServiceUsageState) CalculateDifference(prev *ServiceUsageState) *ServiceUsageStateDiff {
	// resulting difference
	result := &ServiceUsageStateDiff{
		Prev:                 prev,
		Next:                 next,
		ComponentInstantiate: make(map[string]bool),
		ComponentDestruct:    make(map[string]bool)}

	result.calculateDifferenceOnComponentLevel()

	return result
}

// On a service level -- see which keys appear and disappear
func (result *ServiceUsageStateDiff) printDifferenceOnServicesLevel() {

	// High-level service resolutions in prev
	pMap := make(map[string]map[string]string)
	if result.Prev.Dependencies != nil {
		for _, deps := range result.Prev.Dependencies.Dependencies {
			for _, d := range deps {
				if pMap[d.UserID] == nil {
					pMap[d.UserID] = make(map[string]string)
				}
				pMap[d.UserID][d.Service] = d.ResolvesTo
			}
		}
	}

	// High-level service resolutions in next
	cMap := make(map[string]map[string]string)
	for _, deps := range result.Next.Dependencies.Dependencies {
		for _, d := range deps {
			if cMap[d.UserID] == nil {
				cMap[d.UserID] = make(map[string]string)
			}
			cMap[d.UserID][d.Service] = d.ResolvesTo
		}
	}

	// map of all user keys
	uKeys := make(map[string]bool)
	for k := range pMap {
		uKeys[k] = true
	}
	for k := range cMap {
		uKeys[k] = true
	}

	// Printable description
	textMap := make(map[string][]string)

	// Go over all users
	for userID := range uKeys {
		sPrev := pMap[userID]
		sNext := cMap[userID]

		// merge all the service keys
		sKeys := make(map[string]bool)
		for s := range sPrev {
			sKeys[s] = true
		}
		for s := range sNext {
			sKeys[s] = true
		}

		// Process all additions
		for s := range sKeys {
			_, inPrev := sPrev[s]
			_, inNext := sNext[s]
			if !inPrev && inNext {
				textMap[userID] = append(textMap[userID], fmt.Sprintf("[+] %s (%s)", s, cMap[userID][s]))
			}
		}

		// Process all deletions
		for s := range sKeys {
			_, inPrev := sPrev[s]
			_, inNext := sNext[s]
			if inPrev && !inNext {
				textMap[userID] = append(textMap[userID], fmt.Sprintf("[-] %s", s))
			}
		}
	}

	// Print
	printed := false
	for userID, sKeys := range textMap {
		user := LoadUserByIDFromDir(GetAptomiPolicyDir(), userID)
		fmt.Printf("%s (ID=%s)\n", user.Name, user.ID)
		for _, s := range sKeys {
			fmt.Printf("  %s\n", s)
		}
		printed = true
	}

	if (!printed) {
		fmt.Println("[*] No changes")
	}
}

// On a component level -- see which allocation keys appear and disappear
func (result *ServiceUsageStateDiff) calculateDifferenceOnComponentLevel() {
	// map of all instances
	allKeys := make(map[string]bool)

	// merge all the keys
	for k := range result.Prev.ResolvedLinks {
		allKeys[k] = true
	}
	for k := range result.Next.ResolvedLinks {
		allKeys[k] = true
	}
	// Go over all the keys and see which one appear and which one disappear
	for k := range allKeys {
		uPrev := result.Prev.ResolvedLinks[k]
		uNext := result.Next.ResolvedLinks[k]

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
			result.ComponentInstantiate[k] = true
		}

		// see if a component needs to be destructed
		if userIdsPrev != nil && userIdsNext == nil {
			result.ComponentDestruct[k] = true
		}

		// see what needs to happen to users
		uPrevIdsMap := toMap(userIdsPrev)
		uNextIdsMap := toMap(userIdsNext)

		// see if a user needs to be detached from a component
		for u := range uPrevIdsMap {
			if !uNextIdsMap[u] {
				result.ComponentDetachUser = append(result.ComponentDetachUser, ServiceUsageUserAction{ComponentKey: k, User: u})
			}
		}

		// see if a user needs to be attached to a component
		for u := range uNextIdsMap {
			if !uPrevIdsMap[u] {
				result.ComponentAttachUser = append(result.ComponentAttachUser, ServiceUsageUserAction{ComponentKey: k, User: u})
			}
		}
	}
}

func toMap(p []string) map[string]bool {
	result := make(map[string]bool)
	for _, s := range p {
		result[s] = true
	}
	return result
}

/*
func (diff ServiceUsageStateDiff) isEmpty() bool {
	if len(diff.ComponentInstantiate) > 0 {
		return false
	}
	if len(diff.ComponentAttachUser) > 0 {
		return false
	}
	if len(diff.ComponentDetachUser) > 0 {
		return false
	}
	if len(diff.ComponentDestruct) > 0 {
		return false
	}
	return true
}
*/

// Print method prints changes onto the screen (i.e. delta - what got added/removed)
func (diff ServiceUsageStateDiff) Print() {
	/*
	if len(diff.ComponentInstantiate) > 0 {
		fmt.Println("New services to instantiate:")
		for k := range diff.ComponentInstantiate {
			_, _, _, componentName := parseServiceUsageKey(k)
			if componentName == componentRootName {
				fmt.Println("[+] " + k)
			}
		}
	}

	if len(diff.ComponentAttachUser) > 0 {
		fmt.Println("Add users to components:")
		for _, cu := range diff.ComponentAttachUser {
			fmt.Println("[+] " + cu.User + " -> " + cu.ComponentKey)
		}
	}

	if len(diff.ComponentDetachUser) > 0 {
		fmt.Println("Delete users from components:")
		for _, cu := range diff.ComponentDetachUser {
			fmt.Println("[-] " + cu.User + " -> " + cu.ComponentKey)
		}
	}

	if len(diff.ComponentDestruct) > 0 {
		fmt.Println("Components to destruct (no usage):")
		for k := range diff.ComponentDestruct {
			fmt.Println("[-] " + k)
		}
	}

	if diff.isEmpty() {
		fmt.Println("[*] No changes to apply")
	}
	*/

	diff.printDifferenceOnServicesLevel();
}

// Apply method applies all changes via executors and saves usage state in Aptomi DB
func (diff ServiceUsageStateDiff) Apply() {
	// TODO: remove
	diff.Next.SaveServiceUsageState()

	// Process destructions in the right order
	for _, key := range diff.Prev.ProcessingOrder {
		// Does it need to be destructed?
		if _, ok := diff.ComponentDestruct[key]; ok {
			serviceName, _ /*contextName*/ , _ /*allocationName*/ , componentName := parseServiceUsageKey(key)
			component := diff.Prev.Policy.Services[serviceName].getComponentsMap()[componentName]
			if component == nil {
				glog.Infof("Destructing service: %s", serviceName)
				// TODO: add processing code
			} else {
				glog.Infof("Destructing component: %s (%s)", component.Name, component.Code)

				if component.Code != nil {
					codeExecutor, err := component.Code.GetCodeExecutor()
					if err != nil {
						glog.Fatal("Error while getting codeExecutor")
					}
					codeExecutor.Destroy(key)
				}
			}
		}
	}

	// Process instantiations in the right order
	for _, key := range diff.Next.ProcessingOrder {
		// Does it need to be instantiated?
		if _, ok := diff.ComponentInstantiate[key]; ok {
			serviceName, contextName, allocationName, componentName := parseServiceUsageKey(key)
			serviceKey := strings.Join([]string{serviceName, contextName, allocationName, "root"}, "#")
			component := diff.Next.Policy.Services[serviceName].getComponentsMap()[componentName]
			labels := diff.Next.ResolvedLinks[key].CalculatedLabels

			if component == nil {
				glog.Infof("Instantiating service: %s (%s)", serviceName, key)
				// TODO: add processing code
			} else {
				glog.Infof("Instantiating component: %s (%s)", component.Name, key)

				if component.Code != nil {
					codeExecutor, err := component.Code.GetCodeExecutor()
					if err != nil {
						glog.Fatal("Error while getting codeExecutor")
					}
					dependencies := diff.Next.ComponentInstanceMap[serviceKey]
					err = codeExecutor.Install(key, labels, dependencies)
					if err != nil {
						glog.Fatal("Failed install", err)
					}
				}
			}
		}
	}

	// save new state
	diff.Next.SaveServiceUsageState()
}
