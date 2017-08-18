package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/log"
	log "github.com/Sirupsen/logrus"
)

// Endpoints returns map from key to map from port type to url for all services
func (state *ServiceUsageState) Endpoints(filterUserID string) map[string]map[string]string {
	result := make(map[string]map[string]string)

	for _, key := range state.GetResolvedData().ComponentProcessingOrder {
		if _, ok := result[key]; ok {
			continue
		}

		instance := state.GetResolvedData().ComponentInstanceMap[key]
		used := filterUserID == ""
		for dependencyID := range instance.DependencyIds {
			userID := state.Policy.Dependencies.DependenciesByID[dependencyID].UserID
			if userID == filterUserID {
				used = true
				break
			}
		}
		if !used {
			continue
		}

		component := state.Policy.Services[instance.Key.ServiceName].GetComponentsMap()[instance.Key.ComponentName]
		if component != nil && component.Code != nil {
			codeExecutor, err := GetCodeExecutor(component.Code, key, state.GetResolvedData().ComponentInstanceMap[key].CalculatedCodeParams, state.Policy.Clusters)
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
