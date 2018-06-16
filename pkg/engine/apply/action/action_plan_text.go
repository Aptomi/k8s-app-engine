package action

import (
	"sort"
	"strings"

	"github.com/Aptomi/aptomi/pkg/util"
)

// PlanAsText is a plan of actions, represented as text
type PlanAsText struct {
	Actions []util.NestedParameterMap
}

// NewPlanAsText returns new PlanAsText
func NewPlanAsText() *PlanAsText {
	return &PlanAsText{}
}

type keySorter []util.NestedParameterMap

func (ks keySorter) Len() int {
	return len(ks)
}

func (ks keySorter) Swap(i, j int) {
	ks[i], ks[j] = ks[j], ks[i]
}

func (ks keySorter) Less(i, j int) bool {
	return ks[i]["pretty"].(string) < ks[j]["pretty"].(string)
}

// ToString returns human-readable version of the plan
func (t *PlanAsText) String() string {
	// sort actions by pretty text
	sort.Sort(keySorter(t.Actions))

	// action map
	actionDescriptionMap := map[string]string{
		"[+]": "Create Instances",
		"[-]": "Destroy Instances",
		"[*]": "Update Instances",
		"[>]": "Add Consumers",
		"[<]": "Remove Consumers",
		"[@]": "Query Endpoints",
	}

	// combine actions into a string
	result := ""
	actionDescriptionPrev := ""
	for _, pMap := range t.Actions {
		if _, ok := pMap["prettyOmit"]; ok {
			// do not print lines which are indicated as "prettyOmit" by DescribeChanges() in actions
			continue
		}
		if strings.Contains(pMap["key"].(string), "#root") {
			// do not print lines corresponding to root components
			continue
		}

		// get pretty string
		pretty := pMap["pretty"].(string)

		// add action category if needed
		actionDescription := actionDescriptionMap[pretty[:3]]
		if actionDescriptionPrev != actionDescription {
			actionDescriptionPrev = actionDescription
			result += actionDescription + "\n"
		}

		// add pretty action
		result += "  " + pretty + "\n"
	}
	return result
}
