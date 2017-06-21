package slinga

import (
	log "github.com/Sirupsen/logrus"
)

// Endpoints returns map from key to map from port type to url for all services
func (state *ServiceUsageState) Endpoints(filterUserId string) map[string]map[string]string {
	result := make(map[string]map[string]string)

	for _, key := range state.GetResolvedUsage().ComponentProcessingOrder {
		if _, ok := result[key]; ok {
			continue
		}

		instance := state.GetResolvedUsage().ComponentInstanceMap[key]
		used := filterUserId == ""
		for _, dependencyID := range instance.DependencyIds {
			userID := state.Dependencies.DependenciesByID[dependencyID].UserID
			if userID == filterUserId {
				used = true
				break
			}
		}
		if !used {
			continue
		}

		serviceName, _ /*contextName*/, _ /*allocationName*/, componentName := ParseServiceUsageKey(key)
		component := state.Policy.Services[serviceName].getComponentsMap()[componentName]
		if component != nil && component.Code != nil {
			codeExecutor, err := component.Code.GetCodeExecutor(key, state.GetResolvedUsage().ComponentInstanceMap[key].CalculatedCodeParams, state.Policy.Clusters)
			if err != nil {
				debug.WithFields(log.Fields{
					"key":   key,
					"error": err,
				}).Panic("Unable to get CodeExecutor")
			}
			endpoints, err := codeExecutor.Endpoints()
			if err != nil {
				debug.WithFields(log.Fields{
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
