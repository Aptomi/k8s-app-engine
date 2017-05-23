package slinga

import (
	"fmt"
	"github.com/golang/glog"
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
}

// CalculateDifference calculates difference between two given usage states
func (next *ServiceUsageState) CalculateDifference(prev *ServiceUsageState) *ServiceUsageStateDiff {
	// resulting difference
	result := &ServiceUsageStateDiff{
		Prev:                 prev,
		Next:                 next,
		ComponentInstantiate: make(map[string]bool),
		ComponentDestruct:    make(map[string]bool)}

	// map of all instances
	allKeys := make(map[string]bool)

	// merge all the keys
	for k := range prev.ResolvedLinks {
		allKeys[k] = true
	}
	for k := range next.ResolvedLinks {
		allKeys[k] = true
	}

	// go over all the keys and see which one appear and which one disappear
	for k := range allKeys {
		uPrev := prev.ResolvedLinks[k]
		uNext := next.ResolvedLinks[k]

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

	return result
}

func toMap(p []string) map[string]bool {
	result := make(map[string]bool)
	for _, s := range p {
		result[s] = true
	}
	return result
}

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

// Print method prints changes onto the screen (i.e. delta - what got added/removed)
func (diff ServiceUsageStateDiff) Print() {
	if len(diff.ComponentInstantiate) > 0 {
		fmt.Println("New components to instantiate:")
		for k := range diff.ComponentInstantiate {
			fmt.Println("[+] " + k)
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
}

// Apply method applies all changes via executors and saves usage state in Aptomi DB
func (diff ServiceUsageStateDiff) Apply() {
	// TODO: remove
	diff.Next.SaveServiceUsageState()

	// Process destructions in the right order
	for _, key := range diff.Prev.ProcessingOrder {
		// Does it need to be destructed?
		if _, ok := diff.ComponentDestruct[key]; ok {
			serviceName, _ /*contextName*/, _ /*allocationName*/, componentName := parseServiceUsageKey(key)
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
			serviceName, _ /*contextName*/, _ /*allocationName*/, componentName := parseServiceUsageKey(key)
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
					codeExecutor.Install(key, labels)
				}
			}
		}
	}

	// save new state
	diff.Next.SaveServiceUsageState()
}
