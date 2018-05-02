package action

// ApplyResult is a result of applying actions
type ApplyResult struct {
	Success int
	Failed  int
	Skipped int
}

// Plan is a plan of actions
type Plan struct {
	// NodeMap is a map from key to a graph of actions, which must to be executed in order to get from actual state to
	// desired state. Key in the map corresponds to component instance keys.
	NodeMap map[string]*GraphNode

	// RootNodes determines root nodes in the action graph, from which the apply should start.
	RootNodes map[string]bool
}

// NewPlan creates a new Plan
func NewPlan() *Plan {
	return &Plan{
		NodeMap:   make(map[string]*GraphNode),
		RootNodes: make(map[string]bool),
	}
}

// GetActionGraphNode returns an action graph node for a given component instance key
func (plan *Plan) GetActionGraphNode(key string) *GraphNode {
	result, ok := plan.NodeMap[key]
	if !ok {
		result = NewGraphNode(key)
		plan.NodeMap[key] = result
	}
	return result
}

// AddRootNode adds a new root node in the action graph
func (plan *Plan) AddRootNode(key string) {
	plan.RootNodes[key] = true
}

// Apply applies the action plan. It may call fn in multiple go routines, executing the plan in parallel
func (plan *Plan) Apply(fn ApplyFunction) *ApplyResult {
	was := make(map[string]bool)
	wasError := make(map[string]error)
	result := &ApplyResult{}
	for key := range plan.RootNodes {
		_ = plan.applyNode(key, fn, was, wasError, result)
	}
	return result
}

// TODO: change implementation from BFS to DFS (apply in waves for parallelism), https://github.com/Aptomi/aptomi/issues/310
func (plan *Plan) applyNode(key string, fn ApplyFunction, was map[string]bool, wasError map[string]error, result *ApplyResult) error {
	// see if we've been here already
	if was[key] {
		return wasError[key]
	}
	was[key] = true

	// locate the node
	node := plan.NodeMap[key]

	// run all 'before' actions. if one of them fails, don't continue
	for _, beforeNode := range node.Before {
		err := plan.applyNode(beforeNode.Key, fn, was, wasError, result)
		if err != nil {
			return err
		}
	}

	// run all 'main' actions. if one of them fails, don't continue
	for idx, action := range node.Actions {
		err := fn(action)
		if err != nil {
			result.Failed++
			result.Skipped += len(node.Actions) - idx - 1
			return err
		}
		result.Success++
	}

	return nil
}
