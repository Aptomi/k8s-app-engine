package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	"sort"
	. "github.com/Frostman/aptomi/pkg/slinga/language"
)

// SummaryView represents summary view that we show on the home page
type SummaryView struct {
	userID string
	state  slinga.ServiceUsageState
	users  GlobalUsers
}

// NewObjectView creates a new SummaryView
func NewSummaryView(userID string, state slinga.ServiceUsageState, users GlobalUsers) SummaryView {
	return SummaryView{
		userID: userID,
		state:  state,
		users:  users,
	}
}

// GetData returns data for a given view
func (view SummaryView) GetData() interface{} {
	return map[string]interface{}{
		"globalDependencies": view.getGlobalDependenciesData(),
		"globalRules":        view.getGlobalRulesData(),
		"servicesOwned":      view.getServicesOwned(),
		"servicesUsing":      view.getServicesUsing(),
	}
}

func (view SummaryView) getGlobalDependenciesData() interface{} {
	// table only exists for global ops people
	result := lineEntryList{}
	if !view.users.Users[view.userID].IsGlobalOps() {
		return result
	}
	for _, dependency := range view.state.Dependencies.DependenciesByID {
		entry := lineEntry{
			"resolved":     dependency.Resolved,
			"userName":     view.users.Users[dependency.UserID].Name,
			"serviceName":  dependency.Service,
			"context":      view.getResolvedContextNameByDep(dependency),
			"cluster":      view.getResolvedClusterNameByDep(dependency),
			"stats":        view.getDependencyStats(dependency),
			"dependencyId": dependency.ID,
			"id":           view.users.Users[dependency.UserID].Name, // entries will be sorted by ID
		}
		result = append(result, entry)
	}
	sort.Sort(result)
	return result
}

func (view SummaryView) getGlobalRulesData() interface{} {
	// table only exists for global ops people
	result := lineEntryList{}
	if !view.users.Users[view.userID].IsGlobalOps() {
		return result
	}
	for _, ruleList := range view.state.Policy.Rules.Rules {
		for _, rule := range ruleList {
			entry := lineEntry{
				"ruleName":   rule.Name,
				"ruleObject": rule.FilterServices,
				"appliedTo":  view.getRuleAppliedTo(rule),
				// currently we're only matching users by labels (for demo with rules w/o any other filters)
				"matchedUsers": view.getRuleMatchedUsers(rule),
				"conditions":   rule.DescribeConditions(),
				"actions":      rule.DescribeActions(),
				"id":           rule.Name, // entries will be sorted by ID
			}
			result = append(result, entry)
		}
	}
	sort.Sort(result)
	return result
}

func (view SummaryView) getServicesOwned() interface{} {
	result := lineEntryList{}
	for _, service := range view.state.Policy.Services {
		if service.Owner == view.userID {
			// if I own this service, let's find all its instances
			instanceMap := make(map[string]bool)
			for key, instance := range view.state.ResolvedData.ComponentInstanceMap {
				if instance.Resolved {
					serviceName, _, _, componentName := slinga.ParseServiceUsageKey(key)
					if serviceName == service.Name && componentName == slinga.ComponentRootName {
						instanceMap[key] = true
					}
				}
			}

			// Add info about every allocated instance
			for key := range instanceMap {
				instance := view.state.ResolvedData.ComponentInstanceMap[key]
				entry := lineEntry{
					"serviceName": service.Name,
					"context":     view.getResolvedContextNameByInst(instance),
					"cluster":     view.getResolvedClusterNameByInst(instance),
					"stats":       view.getInstanceStats(instance),
					"id":          getWebIDByComponentKey(key), // entries will be sorted by ID
				}
				result = append(result, entry)
			}
		}
	}
	sort.Sort(result)
	return result
}

func (view SummaryView) getServicesUsing() interface{} {
	result := lineEntryList{}
	for _, dependency := range view.state.Dependencies.DependenciesByID {
		if dependency.UserID == view.userID {
			entry := lineEntry{
				"resolved":     dependency.Resolved,
				"serviceName":  dependency.Service,
				"context":      view.getResolvedContextNameByDep(dependency),
				"cluster":      view.getResolvedClusterNameByDep(dependency),
				"stats":        view.getDependencyStats(dependency),
				"dependencyId": dependency.ID,
				"id":           dependency.ID, // entries will be sorted by ID
			}
			result = append(result, entry)
		}
	}
	sort.Sort(result)
	return result
}
