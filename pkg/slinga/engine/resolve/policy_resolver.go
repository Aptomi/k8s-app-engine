package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
	. "github.com/Aptomi/aptomi/pkg/slinga/util"
)

/*
	Core engine for policy resolution
	- takes policy an an input
	- calculates PolicyResolution as an output
*/

type PolicyResolver struct {
	/*
		Input objects
	*/

	// Policy
	policy *PolicyNamespace

	// User loader
	userLoader UserLoader

	/*
		Cache
	*/

	// Expression cache
	expressionCache expression.ExpressionCache

	// Template cache
	templateCache template.TemplateCache

	/*
		Calculated objects
	*/

	// Reference to the calculated PolicyResolution
	resolution *PolicyResolution

	// Buffered event log - gets populated during policy resolution
	eventLog *EventLog
}

// NewPolicyResolver creates a new policy resolver
func NewPolicyResolver(policy *PolicyNamespace, userLoader UserLoader) *PolicyResolver {
	return &PolicyResolver{
		policy:          policy,
		userLoader:      userLoader,
		expressionCache: expression.NewExpressionCache(),
		templateCache:   template.NewTemplateCache(),
		resolution:      NewPolicyResolution(),
		eventLog:        NewEventLog(),
	}
}

// ResolveAllDependencies evaluates and resolves all recorded dependencies ("<user> needs <service> with <labels>"), calculating component allocations
func (resolver *PolicyResolver) ResolveAllDependencies() (*PolicyResolution, error) {
	// Run every declared dependency via policy and resolve it
	for _, dependencies := range resolver.policy.Dependencies.DependenciesByService {
		for _, d := range dependencies {
			// resolve dependency via applying policy
			err := resolver.resolveDependency(d)

			// see if there is an error
			if err != nil {
				return nil, err
			}
		}
	}
	return resolver.resolution, nil
}

// Resolves a single dependency and puts resolution data into the overall state of the world
func (resolver *PolicyResolver) resolveDependency(d *language.Dependency) error {

	// create resolution node
	node := resolver.newResolutionNode(d)

	// aggregate logs in the end
	defer func() {
		for _, eventLog := range node.eventLogsCombined {
			resolver.eventLog.Append(eventLog)
		}
	}()

	// recursively resolve everything
	err := resolver.resolveNode(node)
	if err != nil {
		return err
	}

	// add dependency resolution data to the rest of the records
	d.Resolved = node.resolved
	d.ServiceKey = node.serviceKey.GetKey()

	var data *ResolutionData
	if node.resolved {
		data = resolver.resolution.Resolved
	} else {
		data = resolver.resolution.Unresolved
	}

	err = data.AppendData(node.data)
	if err != nil {
		node.eventLog.LogError(err)
		return err
	}

	return nil
}

// Evaluate evaluates and resolves a single dependency ("<user> needs <service> with <labels>") and calculates component allocations
// Returns error only if there is an issue with the policy (e.g. it's malformed)
// Returns nil if there is no error (it may be that nothing was still matched though)
// If you want to check for successful resolution, use node.resolved flag
func (resolver *PolicyResolver) resolveNode(node *resolutionNode) error {
	// Error variable that we will be reusing
	var err error

	// Indicate that we are starting to resolve dependency
	node.objectResolved(node.dependency)
	node.logStartResolvingDependency()

	// Locate the user
	err = node.checkUserExists()
	if err != nil {
		// If consumer is not present, let's just say that this dependency cannot be fulfilled
		return node.cannotResolveInstance(err)
	}
	node.objectResolved(node.user)

	// Locate the service
	node.service, err = node.getMatchedService(resolver.policy)
	if err != nil {
		// Return a policy processing error in case service is not present in policy
		return node.cannotResolveInstance(err)
	}
	node.objectResolved(node.service)

	// Process service and transform labels
	node.labels = node.transformLabels(node.labels, node.service.ChangeLabels)

	// Match the context
	node.context, err = node.getMatchedContext(resolver.policy)
	if err != nil {
		// Return a policy processing error in case of context resolution failure
		return node.cannotResolveInstance(err)
	}

	// If no matching context is found, the dependency cannot be resolved
	if node.context == nil {
		// This is considered a normal scenario (no matching context found), so no error is returned
		return node.cannotResolveInstance(nil)
	}
	node.objectResolved(node.context)

	// Process context and transform labels
	node.labels = node.transformLabels(node.labels, node.context.ChangeLabels)

	// Resolve allocation keys for the context
	node.allocationKeysResolved, err = node.resolveAllocationKeys(resolver.policy)
	if err != nil {
		// Return an error in case of malformed policy or policy processing error
		return node.cannotResolveInstance(err)
	}

	// Create service key
	node.serviceKey = node.createComponentKey(nil)
	node.objectResolved(node.serviceKey)

	// Check if we've been there already
	cycle := ContainsString(node.path, node.serviceKey.GetKey())
	node.path = append(node.path, node.serviceKey.GetKey())
	if cycle {
		err = node.errorServiceCycleDetected()
		return node.cannotResolveInstance(err)
	}

	// Store labels for service
	node.data.RecordLabels(node.serviceKey, node.labels)

	// Store edge (last component instance -> service instance)
	node.data.StoreEdge(node.arrivalKey, node.serviceKey)

	// Now, sort all components in topological order
	componentsOrdered, err := node.sortServiceComponents()
	if err != nil {
		// Return an error in case of failed component topological sort
		return node.cannotResolveInstance(err)
	}

	// Iterate over all service components and resolve them recursively
	// Note that discovery variables can refer to other variables announced by dependents in the discovery tree
	for _, node.component = range componentsOrdered {
		// Create key
		node.componentKey = node.createComponentKey(node.component)

		// Store edge (service instance -> component instance)
		node.data.StoreEdge(node.serviceKey, node.componentKey)

		// Calculate and store labels for component
		node.componentLabels = node.transformLabels(node.labels, node.component.ChangeLabels)
		node.data.RecordLabels(node.componentKey, node.componentLabels)

		// Create new map with resolution keys for component
		node.discoveryTreeNode[node.component.Name] = NestedParameterMap{}

		// Calculate and store discovery params
		err := node.calculateAndStoreDiscoveryParams()
		if err != nil {
			return node.cannotResolveInstance(err)
		}

		// Print information that we are starting to resolve dependency (on code, or on service)
		node.logResolvingDependencyOnComponent()

		if node.component.Code != nil {
			// Evaluate code params
			err := node.calculateAndStoreCodeParams()
			if err != nil {
				return node.cannotResolveInstance(err)
			}
		} else if node.component.Service != "" {
			// Create a child node for dependency resolution
			nodeNext := node.createChildNode()

			// Resolve dependency on another service recursively
			err := resolver.resolveNode(nodeNext)

			// Combine event logs
			node.eventLogsCombined = append(node.eventLogsCombined, nodeNext.eventLogsCombined...)

			if err != nil {
				return node.cannotResolveInstance(err)
			}

			// If a sub-dependency has not been fulfilled, then exit
			if !nodeNext.resolved {
				// This is considered a normal scenario (sub-dependency not fulfilled), so no error is returned
				return node.cannotResolveInstance(nil)
			}
		}

		// Record usage of a given component instance
		node.logInstanceSuccessfullyResolved(node.componentKey)
		node.data.RecordResolved(node.componentKey, node.dependency)
	}

	// Mark note as resolved and record usage of a given service instance
	node.resolved = true
	node.logInstanceSuccessfullyResolved(node.serviceKey)
	node.data.RecordResolved(node.serviceKey, node.dependency)

	return nil
}
