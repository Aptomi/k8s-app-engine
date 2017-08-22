package visibility

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
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
	instance := view.state.ResolvedData.ComponentInstanceMap[dependency.ServiceKey]
	return view.getResolvedContextNameByInst(instance)
}

func (view SummaryView) getRuleAppliedTo(rule *Rule) string {
	// TODO: complete
	return "-1 instances"
}

func (view SummaryView) getRuleMatchedUsers(rule *Rule) []*User {
	matchedUsers := make([]*User, 0)

	for _, user := range view.state.GetUserLoader().LoadUsersAll().Users {
		match, err := rule.MatchUser(user)
		if err != nil {
			// TODO: we probably need to handle this error better here
			panic(err)
		}
		if match {
			matchedUsers = append(matchedUsers, user)
		}
	}

	return matchedUsers
}

func (view SummaryView) getInstanceStats(instance *engine.ComponentInstance) string {
	runningTime := NewTimeDiff(view.state.ResolvedData.ComponentInstanceMap[instance.Key.GetKey()].GetRunningTime()).Humanize()
	return fmt.Sprintf("%s", runningTime)
}

func (view SummaryView) getResolvedClusterNameByInst(instance *engine.ComponentInstance) string {
	return view.state.ResolvedData.ComponentInstanceMap[instance.Key.GetKey()].CalculatedLabels.Labels["cluster"]
}

func (view SummaryView) getResolvedContextNameByInst(instance *engine.ComponentInstance) string {
	return instance.Key.ContextNameWithKeys
}

func getWebIDByComponentKey(key string) string {
	return strings.Replace(key, "#", ".", -1)
}

func getWebComponentKeyByID(id string) string {
	return strings.Replace(id, ".", "#", -1)
}
