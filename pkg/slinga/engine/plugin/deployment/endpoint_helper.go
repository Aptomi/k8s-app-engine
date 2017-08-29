package deployment

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	log "github.com/Sirupsen/logrus"
)

// Endpoints returns map from key to map from port type to url for all services
func Endpoints(policy *language.PolicyNamespace, state *resolve.ServiceUsageState, filterUserID string) map[string]map[string]string {
	result := make(map[string]map[string]string)

	for _, key := range state.ResolvedData.ComponentProcessingOrder {
		if _, ok := result[key]; ok {
			continue
		}

		instance := state.ResolvedData.ComponentInstanceMap[key]
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
			codeExecutor, err := GetCodeExecutor(component.Code, key, state.ResolvedData.ComponentInstanceMap[key].CalculatedCodeParams, policy.Clusters)
			if err != nil {
				Debug.WithFields(log.Fields{
					"key":   key,
					"error": err,
				}).Panic("Unable to get CodeExecutor")
			}
			endpoints, err := codeExecutor.Endpoints()
			if err != nil {
				Debug.WithFields(log.Fields{
					"key":   key,
					"error": err,
				}).Panic("Error while getting endpoints")
			}

			if len(endpoints) > 0 {
				result[key] = endpoints
			}
		}
	}

	return result
}
