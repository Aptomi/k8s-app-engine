package action

import (
	"sync"
)

// Plan is a plan of actions
type Plan struct {
	// NodeMap is a map from key to a graph of actions, which must to be executed in order to get from actual state to
	// desired state. Key in the map corresponds to the key of the GraphNode
	NodeMap map[string]*GraphNode
}

// NewPlan creates a new Plan
func NewPlan() *Plan {
	return &Plan{
		NodeMap: make(map[string]*GraphNode),
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

// Apply applies the action plan. It may call fn in multiple go routines, executing the plan in parallel
func (plan *Plan) Apply(fn ApplyFunction, resultUpdater ApplyResultUpdater) *ApplyResult {
	// update total number of actions and start the revision
	resultUpdater.SetTotal(plan.NumberOfActions())

	// apply the plan and calculate result (success/failed/skipped actions)
	plan.applyInternal(fn, resultUpdater)

	// tell results updater that we are done and return the results
	return resultUpdater.Done()
}

// Apply applies the action plan. It may call fn in multiple go routines, executing the plan in parallel
func (plan *Plan) applyInternal(fn ApplyFunction, resultUpdater ApplyResultUpdater) {
	deg := make(map[string]int)
	wasError := make(map[string]error)
	queue := make(chan string, len(plan.NodeMap))
	mutex := &sync.RWMutex{}

	// Initialize all degrees, put 0-degree leaf nodes into the queue
	var wg sync.WaitGroup
	for key := range plan.NodeMap {
		deg[key] = len(plan.NodeMap[key].Before)
		if deg[key] <= 0 {
			queue <- key
		}
		wg.Add(1)
	}

	// Start execution
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		// This will keep running until the queue if not closed
		for key := range queue {
			// Take element off the queue, apply the block of actions and put into queue 0-degree nodes which are waiting on us
			go func(key string) {
				defer wg.Done()
				plan.applyActions(key, fn, queue, deg, wasError, mutex, resultUpdater)
			}(key)
		}
		done.Done()
	}()

	// Wait for all actions to finish
	wg.Wait()

	// Close the channel to ensure that the go routine launched above will exit
	close(queue)

	// Wait for the go routine to finish
	done.Wait()
}

// This function applies a block of actions and updates nodes which are waiting on this node
func (plan *Plan) applyActions(key string, fn ApplyFunction, queue chan string, deg map[string]int, wasError map[string]error, mutex *sync.RWMutex, resultUpdater ApplyResultUpdater) {
	// locate the node
	node := plan.NodeMap[key]

	// run all actions. if one of them fails, the rest won't be executed
	// only run them if all dependent nodes succeeded
	mutex.RLock()
	foundErr := wasError[key]
	mutex.RUnlock()
	for _, action := range node.Actions {
		// if an error happened before, all subsequent actions are getting marked as skipped
		if foundErr != nil {
			// fmt.Println("skipped ", action.GetName())
			resultUpdater.AddSkipped()
		} else {
			// Otherwise, let's run the action and see if it failed or not
			err := fn(action)
			if err != nil {
				// fmt.Println("failed ", action.GetName())
				resultUpdater.AddFailed()
				foundErr = err
			} else {
				// fmt.Println("success ", action.GetName())
				resultUpdater.AddSuccess()
			}
		}
	}

	// mark our node as failed, if we encountered an error
	if foundErr != nil {
		mutex.Lock()
		wasError[key] = foundErr
		mutex.Unlock()
	}

	// decrement degrees of nodes which are waiting on us
	for _, prevNode := range plan.NodeMap[node.Key].BeforeRev {
		mutex.Lock()
		deg[prevNode.Key]--
		if deg[prevNode.Key] < 0 {
			panic("negative node degree while applying actions in parallel")
		}
		if deg[prevNode.Key] == 0 {
			queue <- prevNode.Key
		}
		mutex.Unlock()
		if foundErr != nil {
			// Mark prev nodes failed too
			mutex.Lock()
			wasError[prevNode.Key] = foundErr
			mutex.Unlock()
		}
	}

}

// NumberOfActions returns the total number of actions that is expected to be executed in the whole action graph
func (plan *Plan) NumberOfActions() uint32 {
	resultUpdater := NewApplyResultUpdaterImpl()

	// apply the plan and calculate result (success/failed/skipped actions)
	plan.applyInternal(Noop(), resultUpdater)

	// return the number of success actions (all of them will be success due to Noop() action)
	return resultUpdater.Result.Success
}

// AsText returns the action plan as array of actions, each represented as text via NestedParameterMap
func (plan *Plan) AsText() *PlanAsText {
	result := NewPlanAsText()

	// apply the plan and capture actions as text
	plan.applyInternal(WrapSequential(func(act Base) error {
		result.Actions = append(result.Actions, act.DescribeChanges())
		return nil
	}), NewApplyResultUpdaterImpl())

	return result
}
