package visibility

import (
	"fmt"
	"github.com/Frostman/aptomi/pkg/slinga"
)

type dependencyNode struct {
	dependency *slinga.Dependency
}

func newDependencyNode(dependency *slinga.Dependency) graphNode {
	return dependencyNode{dependency: dependency}
}

func (n dependencyNode) getIDPrefix() string {
	return "dep-"
}

func (n dependencyNode) getGroup() string {
	return "dependency"
}

func (n dependencyNode) getID() string {
	return fmt.Sprintf("%s%s", n.getIDPrefix(), n.dependency.ID)
}

func (n dependencyNode) isItMyID(id string) string {
	return cutPrefixOrEmpty(id, n.getIDPrefix())
}

func (n dependencyNode) getLabel() string {
	return slinga.LoadUserByIDFromDir(slinga.GetAptomiBaseDir(), n.dependency.UserID).Name
}

func (n dependencyNode) getEdgeLabel(dst graphNode) string {
	return ""
}

func (n dependencyNode) getDetails(id string, state slinga.ServiceUsageState) interface{} {
	return state.Dependencies.DependenciesByID[id]
}
