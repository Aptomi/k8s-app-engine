package resolve

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/util"
)

// PolicyResolution contains resolution data for the policy. It essentially represents the desired state calculated
// by policy resolver. It contains a calculated map of component instances with their data, information about
// resolved service consumption declarations, as well as processing order to components in which they have to be
// instantiated/updated/deleted.
type PolicyResolution struct {
	// Resolved component instances: componentKey -> componentInstance
	ComponentInstanceMap map[string]*ComponentInstance

	// Resolved dependencies: dependencyID -> serviceKey
	DependencyInstanceMap map[string]string

	// Resolved component processing order in which components/services have to be processed
	componentProcessingOrderHas map[string]bool
	ComponentProcessingOrder    []string
}

// NewPolicyResolution creates new empty PolicyResolution
func NewPolicyResolution() *PolicyResolution {
	return &PolicyResolution{
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		DependencyInstanceMap:       make(map[string]string),
		componentProcessingOrderHas: make(map[string]bool),
		ComponentProcessingOrder:    []string{},
	}
}

// GetComponentInstanceEntry retrieves a component instance entry by key, or creates an new entry if it doesn't exist
func (resolution *PolicyResolution) GetComponentInstanceEntry(cik *ComponentInstanceKey) *ComponentInstance {
	key := cik.GetKey()
	if _, ok := resolution.ComponentInstanceMap[key]; !ok {
		resolution.ComponentInstanceMap[key] = newComponentInstance(cik)
	}
	return resolution.ComponentInstanceMap[key]
}

// RecordResolved takes a component instance and adds a new dependency record into it
func (resolution *PolicyResolution) RecordResolved(cik *ComponentInstanceKey, dependency *lang.Dependency, ruleResult *lang.RuleActionResult) {
	instance := resolution.GetComponentInstanceEntry(cik)
	instance.addDependency(object.GetKey(dependency))
	instance.addRuleInformation(ruleResult)
	resolution.recordProcessingOrder(cik)
}

// Record processing order for component instance
func (resolution *PolicyResolution) recordProcessingOrder(cik *ComponentInstanceKey) {
	key := cik.GetKey()
	if !resolution.componentProcessingOrderHas[key] {
		resolution.componentProcessingOrderHas[key] = true
		resolution.ComponentProcessingOrder = append(resolution.ComponentProcessingOrder, key)
	}
}

// RecordCodeParams stores calculated code params for component instance
func (resolution *PolicyResolution) RecordCodeParams(cik *ComponentInstanceKey, codeParams util.NestedParameterMap) error {
	return resolution.GetComponentInstanceEntry(cik).addCodeParams(codeParams)
}

// RecordDiscoveryParams stores calculated discovery params for component instance
func (resolution *PolicyResolution) RecordDiscoveryParams(cik *ComponentInstanceKey, discoveryParams util.NestedParameterMap) error {
	return resolution.GetComponentInstanceEntry(cik).addDiscoveryParams(discoveryParams)
}

// RecordLabels stores calculated labels for component instance
func (resolution *PolicyResolution) RecordLabels(cik *ComponentInstanceKey, labels *lang.LabelSet) {
	resolution.GetComponentInstanceEntry(cik).addLabels(labels)
}

// StoreEdge stores incoming/outgoing graph edges for component instance for observability and reporting
func (resolution *PolicyResolution) StoreEdge(src *ComponentInstanceKey, dst *ComponentInstanceKey) {
	// Arrival key can be empty at the very top of the recursive function in engine, so let's check for that
	if src != nil && dst != nil {
		resolution.GetComponentInstanceEntry(src).addEdgeOut(dst.GetKey())
		resolution.GetComponentInstanceEntry(dst).addEdgeIn(src.GetKey())
	}
}

// AppendData appends data to the current PolicyResolution record by aggregating data over component instances
func (resolution *PolicyResolution) AppendData(ops *PolicyResolution) error {
	for _, instance := range ops.ComponentInstanceMap {
		err := resolution.GetComponentInstanceEntry(instance.Metadata.Key).appendData(instance)
		if err != nil {
			return err
		}
	}
	for key := range ops.ComponentInstanceMap {
		resolution.recordProcessingOrder(ops.ComponentInstanceMap[key].Metadata.Key)
	}
	return nil
}
