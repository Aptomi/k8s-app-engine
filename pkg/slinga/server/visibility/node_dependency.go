package visibility

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
)

type dependencyNode struct {
	dependency *Dependency
	short      bool
	userLoader UserLoader
}

func newDependencyNode(dependency *Dependency, short bool, userLoader UserLoader) graphNode {
	return dependencyNode{
		dependency: dependency,
		short:      short,
		userLoader: userLoader,
	}
}

func (n dependencyNode) getIDPrefix() string {
	return "dep-"
}

func (n dependencyNode) getGroup() string {
	if n.short {
		return "dependencyShort"
	}
	if n.dependency.Resolved {
		return "dependencyLongResolved"
	}
	return "dependencyLongNotResolved"
}

func (n dependencyNode) getID() string {
	return fmt.Sprintf("%s%s", n.getIDPrefix(), n.dependency.ID)
}

func (n dependencyNode) isItMyID(id string) string {
	return cutPrefixOrEmpty(id, n.getIDPrefix())
}

func (n dependencyNode) getLabel() string {
	userName := n.userLoader.LoadUserByID(n.dependency.UserID).Name
	if n.short {
		// for service owner view, don't display much other than a user name
		return userName
	}
	// for consumer view - display full dependency info "user name -> service"
	return fmt.Sprintf("%s \u2192 %s", userName, n.dependency.Service)
}

func (n dependencyNode) getEdgeLabel(dst graphNode) string {
	return ""
}

func (n dependencyNode) getDetails(id string, state engine.ServiceUsageState) interface{} {
	return state.Policy.Dependencies.DependenciesByID[id]
}
