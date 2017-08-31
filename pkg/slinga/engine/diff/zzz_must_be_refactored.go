package diff

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/Sirupsen/logrus"
)

// RevisionSummary returns integer counts for all policy objects
type RevisionSummary struct {
	Services     int
	Contexts     int
	Clusters     int
	Rules        int
	UserText     string
	Dependencies int
}

// compare everything, but not users. users are external to us
func (summary RevisionSummary) equal(that RevisionSummary) bool {
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
func GetSummary(revision *resolve.Revision) RevisionSummary {
	if revision.Policy == nil {
		return RevisionSummary{}
	}
	return RevisionSummary{
		util.CountElements(revision.Policy.Services),
		util.CountElements(revision.Policy.Contexts),
		util.CountElements(revision.Policy.Clusters),
		util.CountElements(revision.Policy.Rules.Rules),
		revision.UserLoader.Summary(),
		util.CountElements(revision.Policy.Dependencies.DependenciesByID),
	}
}

// On a service level -- see which keys appear and disappear
func (diff *RevisionDiff) writeDifferenceOnServicesLevel(log *logrus.Logger) {

	log.Println("[Services]")

	// High-level service resolutions in prev (user -> serviceName -> serviceKey -> count)
	pMap := make(map[string]map[string]map[string]int)
	if diff.Prev.Policy != nil {
		for _, deps := range diff.Prev.Policy.Dependencies.DependenciesByService {
			for _, d := range deps {
				// Make sure to check for the case when dependency has not been resolved (e.g. no matching context found or error)
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
	for _, deps := range diff.Next.Policy.Dependencies.DependenciesByService {
		for _, d := range deps {
			// Make sure to check for the case when dependency has not been resolved (e.g. no matching context found or error)
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
		user := diff.Next.UserLoader.LoadUserByID(userID)
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
func (diff RevisionDiff) writeSummary(log *logrus.Logger) {
	prev := GetSummary(diff.Prev)
	next := GetSummary(diff.Next)
	log.Println("[Policy]")
	log.Println(getSummaryLine("Services", prev.Services, next.Services))
	log.Println(getSummaryLine("Contexts", prev.Contexts, next.Contexts))
	log.Println(getSummaryLine("Dependencies", prev.Dependencies, next.Dependencies))
	log.Println(getSummaryLine("Clusters", prev.Clusters, next.Clusters))
	log.Println(getSummaryLine("Rules", prev.Rules, next.Rules))
	log.Println(getSummaryLineAsText("Users", prev.UserText, next.UserText))
}

// ShouldGenerateNewRevision became one of the key methods.
// If it returns false, then new run and new revision will not be generated
func (diff *RevisionDiff) ShouldGenerateNewRevision() bool {
	// TODO: this method should take into account:
	// - policy (input objects)
	// - resolution data (output objects)
	// - plugin data
	// ...

	if !GetSummary(diff.Next).equal(GetSummary(diff.Prev)) {
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

	return false
}
