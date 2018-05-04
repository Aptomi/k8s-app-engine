package action

import (
	"sync"
	"sync/atomic"
)

// ApplyResult is a result of applying actions
type ApplyResult struct {
	Success uint32
	Failed  uint32
	Skipped uint32
}

// Plan is a plan of actions
type Plan struct {
	// NodeMap is a map from key to a graph of actions, which must to be executed in order to get from actual state to
	// desired state. Key in the map corresponds to component instance keys.
	NodeMap map[string]*GraphNode

	// LeafNodes determines leaf nodes in the action graph (i.e. without dependencies), from which the apply should start.
	LeafNodes map[string]bool
}

// NewPlan creates a new Plan
func NewPlan() *Plan {
	return &Plan{
		NodeMap:   make(map[string]*GraphNode),
		LeafNodes: make(map[string]bool),
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

// AddLeafNode adds a new root node in the action graph
func (plan *Plan) AddLeafNode(key string) {
	plan.LeafNodes[key] = true
}

// Apply applies the action plan. It may call fn in multiple go routines, executing the plan in parallel
func (plan *Plan) Apply(fn ApplyFunction) *ApplyResult {
	deg := make(map[string]int)
	wasError := make(map[string]error)
	queue := make(chan string, len(plan.NodeMap))
	mutex := &sync.RWMutex{}
	result := &ApplyResult{}

	// Start from the leaf nodes
	for key := range plan.LeafNodes {
		queue <- key
	}

	// Initialize all degrees
	var wg sync.WaitGroup
	for key := range plan.NodeMap {
		deg[key] = len(plan.NodeMap[key].Before)
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
				plan.applyActions(key, fn, queue, deg, wasError, mutex, result)
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

	return result
}

// This function applies a block of actions and updates nodes which are waiting on this node
func (plan *Plan) applyActions(key string, fn ApplyFunction, queue chan string, deg map[string]int, wasError map[string]error, mutex *sync.RWMutex, result *ApplyResult) {
	// locate the node
	node := plan.NodeMap[key]

	// run all actions. if one of them fails, the rest won't be executed
	// only run them if all dependent nodes succeeded
	mutex.RLock()
	foundErr := wasError[key]
	mutex.RUnlock()
	if foundErr == nil {
		for _, action := range node.Actions {
			// if an error happened before, all subsequent actions are getting marked as skipped
			if foundErr != nil {
				atomic.AddUint32(&result.Skipped, 1)
			} else {
				// Otherwise, let's run the action and see if it failed or not
				err := fn(action)
				if err != nil {
					atomic.AddUint32(&result.Failed, 1)
					foundErr = err
				} else {
					atomic.AddUint32(&result.Success, 1)
				}
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
	return plan.Apply(Noop()).Success
}
