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

func (n dependencyNode) getID() string {
	return fmt.Sprintf("dep-%s", n.dependency.ID)
}

func (n dependencyNode) getLabel() string {
	return n.dependency.UserID
}

func (n dependencyNode) getGroup() string {
	return "dependency"
}

func (n dependencyNode) getEdgeLabel(dst graphNode) string {
	return ""
}
