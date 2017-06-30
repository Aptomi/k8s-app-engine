package visibility

import (
	"fmt"
	"github.com/Frostman/aptomi/pkg/slinga/engine"
	. "github.com/Frostman/aptomi/pkg/slinga/language"
	. "github.com/Frostman/aptomi/pkg/slinga/util"
	"strings"
)

func (view SummaryView) getDependencyStats(dependency *Dependency) string {
	if !dependency.Resolved {
		return "N/A"
	}
	runningTime := NewTimeDiff(view.state.ResolvedData.ComponentInstanceMap[dependency.ServiceKey].GetRunningTime()).Humanize()
	return fmt.Sprintf("%s", runningTime)
}

func (view SummaryView) getResolvedClusterNameByDep(dependency *Dependency) string {
	if !dependency.Resolved {
		return "N/A"
	}
	return view.state.ResolvedData.ComponentInstanceMap[dependency.ServiceKey].CalculatedLabels.Labels["cluster"]
}

func (view SummaryView) getResolvedContextNameByDep(dependency *Dependency) string {
	if !dependency.Resolved {
		return "N/A"
	}
	_, context, allocation, _ := engine.ParseServiceUsageKey(dependency.ServiceKey)
	return fmt.Sprintf("%s/%s", context, allocation)
}

func (view SummaryView) getRuleAppliedTo(rule *Rule) string {
	// TODO: complete
	return "-1 instances"
}

func (view SummaryView) getRuleMatchedUsers(rule *Rule) []*User {
	matchedUsers := make([]*User, 0)

	for _, user := range view.state.GetUserLoader().LoadUsersAll().Users {
		if rule.MatchUser(user) {
			matchedUsers = append(matchedUsers, user)
		}
	}

	return matchedUsers
}

func (view SummaryView) getInstanceStats(instance *engine.ComponentInstance) string {
	runningTime := NewTimeDiff(view.state.ResolvedData.ComponentInstanceMap[instance.Key].GetRunningTime()).Humanize()
	return fmt.Sprintf("%s", runningTime)
}

func (view SummaryView) getResolvedClusterNameByInst(instance *engine.ComponentInstance) string {
	return view.state.ResolvedData.ComponentInstanceMap[instance.Key].CalculatedLabels.Labels["cluster"]
}

func (view SummaryView) getResolvedContextNameByInst(instance *engine.ComponentInstance) string {
	_, context, allocation, _ := engine.ParseServiceUsageKey(instance.Key)
	return fmt.Sprintf("%s/%s", context, allocation)
}

func getWebIDByComponentKey(key string) string {
	return strings.Replace(key, "#", ".", -1)
}

func getWebComponentKeyByID(id string) string {
	return strings.Replace(id, ".", "#", -1)
}
