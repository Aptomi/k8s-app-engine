package deployment

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

// Endpoints returns map from key to map from port type to url for all services
func Endpoints(policy *language.PolicyNamespace, resolution *resolve.PolicyResolution, filterUserID string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)

	for key := range resolution.ComponentInstanceMap {
		if _, ok := result[key]; ok {
			continue
		}

		instance := resolution.ComponentInstanceMap[key]
		used := filterUserID == ""
		for dependencyID := range instance.DependencyIds {
			userID := policy.Dependencies.DependenciesByID[dependencyID].UserID
			if userID == filterUserID {
				used = true
				break
			}
		}
		if !used {
			continue
		}

		component := policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]
		if component != nil && component.Code != nil {
			codeExecutor, err := GetCodeExecutor(component.Code, key, resolution.ComponentInstanceMap[key].CalculatedCodeParams, policy.Clusters, eventlog.NewEventLog())
			if err != nil {
				return nil, fmt.Errorf("Unable to get CodeExecutor for '%s': %s", key, err.Error())
			}
			endpoints, err := codeExecutor.Endpoints()
			if err != nil {
				return nil, fmt.Errorf("Error while getting endpoints for '%s': %s", key, err.Error())
			}

			if len(endpoints) > 0 {
				result[key] = endpoints
			}
		}
	}

	return result, nil
}
