package visibility

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"strings"
)

func (view SummaryView) getDependencyStats(dependency *lang.Dependency) string {
	/*
		if !dependency.Resolved {
			return "N/A"
		}
	*/
	return "" //NewTimeDiff(view.revision.Resolution.ComponentInstanceMap[dependency.ServiceKey].GetRunningTime()).Humanize()
}

func (view SummaryView) getResolvedClusterNameByDep(dependency *lang.Dependency) string {
	/*
		if !dependency.Resolved {
			return "N/A"
		}
	*/
	return "" //view.revision.Resolution.ComponentInstanceMap[dependency.ServiceKey].CalculatedLabels.Labels[lang.LabelCluster]
}

func (view SummaryView) getResolvedContextNameByDep(dependency *lang.Dependency) string {
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

func (view SummaryView) getRuleAppliedTo(rule *lang.Rule) string {
	// TODO: complete
	return "-1 instances"
}

func (view SummaryView) getRuleMatchedUsers(rule *lang.Rule) []*lang.User {
	matchedUsers := make([]*lang.User, 0)

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
	return "" //view.revision.Resolution.ComponentInstanceMap[instance.Key.GetKey()].CalculatedLabels.Labels[lang.LabelCluster]
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
