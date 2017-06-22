package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	"github.com/Frostman/aptomi/pkg/slinga/time"
	"fmt"
	"strings"
)

func (view SummaryView) getDependencyStats(dependency *slinga.Dependency) string {
	if !dependency.Resolved {
		return "N/A"
	}
	containers := view.getNumberOfContainersByDep(dependency)
	runningTime := time.NewDiff(view.state.ResolvedData.ComponentInstanceMap[dependency.ServiceKey].GetRunningTime()).Humanize()
	return fmt.Sprintf("%d containers/%s running", containers, runningTime)
}

func (view SummaryView) getNumberOfContainersByDep(dependency *slinga.Dependency) int {
	// TODO: return number of containers for a given dependency
	return -1
}

func (view SummaryView) getResolvedClusterNameByDep(dependency *slinga.Dependency) string {
	if !dependency.Resolved {
		return "N/A"
	}
	return view.state.ResolvedData.ComponentInstanceMap[dependency.ServiceKey].CalculatedLabels.Labels["cluster"]
}

func (view SummaryView) getResolvedContextNameByDep(dependency *slinga.Dependency) string {
	if !dependency.Resolved {
		return "N/A"
	}
	_, context, allocation, _ := slinga.ParseServiceUsageKey(dependency.ServiceKey)
	return fmt.Sprintf("%s/%s", context, allocation)
}

func (view SummaryView) getRuleAppliedTo(rule *slinga.Rule) string {
	// TODO: complete
	return "-1 instances"
}

func (view SummaryView) getInstanceStats(instance *slinga.ComponentInstance) string {
	containers := view.getNumberOfContainersByInst(instance)
	runningTime := time.NewDiff(view.state.ResolvedData.ComponentInstanceMap[instance.Key].GetRunningTime()).Humanize()
	return fmt.Sprintf("%d containers/%s running", containers, runningTime)
}

func (view SummaryView) getResolvedClusterNameByInst(instance *slinga.ComponentInstance) string {
	return view.state.ResolvedData.ComponentInstanceMap[instance.Key].CalculatedLabels.Labels["cluster"]
}

func (view SummaryView) getResolvedContextNameByInst(instance *slinga.ComponentInstance) string {
	_, context, allocation, _ := slinga.ParseServiceUsageKey(instance.Key)
	return fmt.Sprintf("%s/%s", context, allocation)
}

func (view SummaryView) getNumberOfContainersByInst(instance *slinga.ComponentInstance) int {
	// TODO: return number of containers for a given instance
	return -1
}

func getWebIDByComponentKey(key string) string {
	return strings.Replace(key, "#", ".", -1)
}

func getWebComponentKeyByID(id string) string {
	return strings.Replace(id, ".", "#", -1)
}
