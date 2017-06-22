package visibility

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	"github.com/Frostman/aptomi/pkg/slinga/time"
	"fmt"
)

func (view SummaryView) getDependencyStats(dependency *slinga.Dependency) string {
	if !dependency.Resolved {
		return "N/A"
	}
	containers := view.getNumberOfContainers(dependency)
	runningTime := time.NewDiff(view.state.ResolvedData.ComponentInstanceMap[dependency.ServiceKey].GetRunningTime()).Humanize()
	return fmt.Sprintf("%d containers/%s running", containers, runningTime)
}

func (view SummaryView) getNumberOfContainers(dependency *slinga.Dependency) int {
	// TODO: return number of containers for a given dependency
	return -1
}

func (view SummaryView) getResolvedClusterName(dependency *slinga.Dependency) string {
	if !dependency.Resolved {
		return "N/A"
	}
	return view.state.ResolvedData.ComponentInstanceMap[dependency.ServiceKey].CalculatedLabels.Labels["cluster"]
}

func (view SummaryView) getResolvedContextName(dependency *slinga.Dependency) string {
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
