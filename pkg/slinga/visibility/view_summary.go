package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	"sort"
)

// SummaryView represents summary view that we show on the home page
type SummaryView struct {
	userID string
	state  slinga.ServiceUsageState
	users  slinga.GlobalUsers
}

// NewObjectView creates a new SummaryView
func NewSummaryView(userID string, state slinga.ServiceUsageState, users slinga.GlobalUsers) SummaryView {
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
			"context":      view.getResolvedContextName(dependency),
			"cluster":      view.getResolvedClusterName(dependency),
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
				"ruleName":     rule.Name,
				"ruleObject":   rule.FilterServices,
				"appliedTo":    view.getRuleAppliedTo(rule),
				"id":           rule.Name, // entries will be sorted by ID
			}
			result = append(result, entry)
		}
	}
	sort.Sort(result)
	return result
}

func (view SummaryView) getServicesOwned() interface{} {
	return "not implemented"
}

func (view SummaryView) getServicesUsing() interface{} {
	return "not implemented"
}
