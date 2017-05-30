package slinga

import (
	log "github.com/Sirupsen/logrus"
)

func (state *ServiceUsageState) Endpoints() map[string]map[string]string {
	result := make(map[string]map[string]string)

	for _, key := range state.ProcessingOrder {
		if _, ok := result[key]; ok {
			continue
		}

		serviceName, _ /*contextName*/, _ /*allocationName*/, componentName := ParseServiceUsageKey(key)
		component := state.Policy.Services[serviceName].getComponentsMap()[componentName]
		if component != nil && component.Code != nil {
			codeExecutor, err := component.Code.GetCodeExecutor(key, component.Code.Metadata, state.ResolvedLinks[key].CalculatedCodeParams, state.Policy.Clusters)
			if err != nil {
				debug.WithFields(log.Fields{
					"key":   key,
					"error": err,
				}).Fatal("Unable to get CodeExecutor")
			}
			endpoints, err := codeExecutor.Endpoints()
			if err != nil {
				debug.WithFields(log.Fields{
					"key":   key,
					"error": err,
				}).Fatal("Error while getting endpoints")
			}

			if len(endpoints) > 0 {
				result[key] = endpoints
			}
		}
	}

	return result
}
