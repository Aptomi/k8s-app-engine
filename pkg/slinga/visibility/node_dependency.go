package visibility

import (
	"fmt"
	"github.com/Frostman/aptomi/pkg/slinga"
)

type dependencyNode struct {
	dependency *slinga.Dependency
	short      bool
}

func newDependencyNode(dependency *slinga.Dependency, short bool) graphNode {
	return dependencyNode{
		dependency: dependency,
		short:      short,
	}
}

func (n dependencyNode) getIDPrefix() string {
	return "dep-"
}

func (n dependencyNode) getGroup() string {
	if n.short {
		return "dependencyShort"
	}
	if len(n.dependency.ResolvesTo) > 0 {
		return "dependencyLongResolved"
	} else {
		return "dependencyLongNotResolved"
	}
}

func (n dependencyNode) getID() string {
	return fmt.Sprintf("%s%s", n.getIDPrefix(), n.dependency.ID)
}

func (n dependencyNode) isItMyID(id string) string {
	return cutPrefixOrEmpty(id, n.getIDPrefix())
}

func (n dependencyNode) getLabel() string {
	userName := slinga.LoadUserByIDFromDir(slinga.GetAptomiBaseDir(), n.dependency.UserID).Name
	if n.short {
		// for service owner view, don't display much other than a user name
		return userName
	} else {
		// for consumer view - display full dependency info "user name -> service"
		return fmt.Sprintf("%s \u2192 %s", userName, n.dependency.Service)
	}
}

func (n dependencyNode) getEdgeLabel(dst graphNode) string {
	return ""
}

func (n dependencyNode) getDetails(id string, state slinga.ServiceUsageState) interface{} {
	return state.Dependencies.DependenciesByID[id]
}
