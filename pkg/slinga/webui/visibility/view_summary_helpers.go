package visibility

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"strings"
)

func (view SummaryView) getDependencyStats(dependency *Dependency) string {
	/*
		if !dependency.Resolved {
			return "N/A"
		}
	*/
	return "" //NewTimeDiff(view.revision.Resolution.ComponentInstanceMap[dependency.ServiceKey].GetRunningTime()).Humanize()
}

func (view SummaryView) getResolvedClusterNameByDep(dependency *Dependency) string {
	/*
		if !dependency.Resolved {
			return "N/A"
		}
	*/
	return "" //view.revision.Resolution.ComponentInstanceMap[dependency.ServiceKey].CalculatedLabels.Labels[language.LabelCluster]
}

func (view SummaryView) getResolvedContextNameByDep(dependency *Dependency) string {
	/*
		if !dependency.Resolved {
			return "N/A"
		}
	*/
	return "" /*
		instance := view.revision.Resolution.ComponentInstanceMap[dependency.ServiceKey]
		return view.getResolvedContextNameByInst(instance)
	*/
}

func (view SummaryView) getRuleAppliedTo(rule *Rule) string {
	// TODO: complete
	return "-1 instances"
}

func (view SummaryView) getRuleMatchedUsers(rule *Rule) []*User {
	matchedUsers := make([]*User, 0)

	/*
		for _, user := range view.revision.UserLoader.LoadUsersAll().Users {
			match, err := rule.MatchUser(user)
			if err != nil {
				// TODO: we probably need to handle this error better here
				panic(err)
			}
			if match {
				matchedUsers = append(matchedUsers, user)
			}
		}
	*/

	return matchedUsers
}

func (view SummaryView) getInstanceStats(instance *resolve.ComponentInstance) string {
	return "" //NewTimeDiff(view.revision.Resolution.ComponentInstanceMap[instance.Key.GetKey()].GetRunningTime()).Humanize()
}

func (view SummaryView) getResolvedClusterNameByInst(instance *resolve.ComponentInstance) string {
	return "" //view.revision.Resolution.ComponentInstanceMap[instance.Key.GetKey()].CalculatedLabels.Labels[language.LabelCluster]
}

func (view SummaryView) getResolvedContextNameByInst(instance *resolve.ComponentInstance) string {
	return instance.Metadata.Key.ContextNameWithKeys
}

func getWebIDByComponentKey(key string) string {
	return strings.Replace(key, "#", ".", -1)
}

func getWebComponentKeyByID(id string) string {
	return strings.Replace(id, ".", "#", -1)
}
