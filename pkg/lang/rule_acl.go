package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

// ACLRule is a struct for defining rules which map users to their roles in Aptomi
type ACLRule = Rule

// ACLRuleObject is an informational data structure with Kind and Constructor for ACLRule
var ACLRuleObject = &object.Info{
	Kind:        "aclrule",
	Versioned:   true,
	Constructor: func() object.Base { return &ACLRule{} },
}

// Allows to define a role which spans across all namespaces (e.g. "domain admin")
const namespaceAll = "*"

// Special label name to add a role to the user
const labelRole = "addrole"

// ACLRole is a struct for defining user roles
// See the list of defined roles below - domain admin, namespace admin, service consumer, and nobody
type ACLRole struct {
	ID         string
	Name       string
	Privileges *Privileges
}

// Privileges defines a set of privileges for a particular role in Aptomi
type Privileges struct {
	// If AllNamespaces is set to true, then those user privileges apply to all namespaces
	// Otherwise it applies to a set of given namespaces
	AllNamespaces bool

	// NamespaceObjects specifies whether or not this role can view/manage a certain object kind within a non-system namespace
	NamespaceObjects map[string]*Privilege

	// GlobalObjects specifies whether or not this role can view/manage a certain object kind within a system namespace
	GlobalObjects map[string]*Privilege
}

// Privilege is a unit of privilege for any single given object
type Privilege struct {
	// View indicates whether or not a user can view an object (R)
	View bool

	// Manage indicates whether or not a user can manage an object, i.e. perform operations (CUD)
	Manage bool
}

// Full access privilege
var fullAccess = &Privilege{
	View:   true,
	Manage: true,
}

// View access privilege
var viewAccess = &Privilege{
	View: true,
}

// No access privilege
var noAccess = &Privilege{}

// Domain admin role
var domainAdmin = &ACLRole{
	ID:   "domain_admin",
	Name: "Domain Admin",
	Privileges: &Privileges{
		AllNamespaces: true,
		NamespaceObjects: map[string]*Privilege{
			ServiceObject.Kind:    fullAccess,
			ContractObject.Kind:   fullAccess,
			DependencyObject.Kind: fullAccess,
			RuleObject.Kind:       fullAccess,
		},
		GlobalObjects: map[string]*Privilege{
			ClusterObject.Kind: fullAccess,
			RuleObject.Kind:    fullAccess,
		},
	},
}

// Namespace admin role
var namespaceAdmin = &ACLRole{
	ID:   "namespace_admin",
	Name: "Namespace Admin",
	Privileges: &Privileges{
		NamespaceObjects: map[string]*Privilege{
			ServiceObject.Kind:    fullAccess,
			ContractObject.Kind:   fullAccess,
			DependencyObject.Kind: fullAccess,
			RuleObject.Kind:       fullAccess,
		},
		GlobalObjects: map[string]*Privilege{
			ClusterObject.Kind: viewAccess,
			RuleObject.Kind:    viewAccess,
		},
	},
}

// Service consumer role
var serviceConsumer = &ACLRole{
	ID:   "service_consumer",
	Name: "Service Consumer",
	Privileges: &Privileges{
		NamespaceObjects: map[string]*Privilege{
			ServiceObject.Kind:    viewAccess,
			ContractObject.Kind:   viewAccess,
			DependencyObject.Kind: fullAccess,
			RuleObject.Kind:       viewAccess,
		},
		GlobalObjects: map[string]*Privilege{
			ClusterObject.Kind: viewAccess,
			RuleObject.Kind:    viewAccess,
		},
	},
}

// Nobody role
var nobody = &ACLRole{
	ID:   "nobody",
	Name: "Nobody",
	Privileges: &Privileges{
		NamespaceObjects: map[string]*Privilege{
			ServiceObject.Kind:    viewAccess,
			ContractObject.Kind:   viewAccess,
			DependencyObject.Kind: viewAccess,
			RuleObject.Kind:       viewAccess,
		},
		GlobalObjects: map[string]*Privilege{
			ClusterObject.Kind: viewAccess,
			RuleObject.Kind:    viewAccess,
		},
	},
}

// ACLRolesOrdered represents the ordered list of ACL roles (from most "powerful" to least "powerful")
var ACLRolesOrdered = map[string]*ACLRole{
	domainAdmin.ID:     domainAdmin,
	namespaceAdmin.ID:  namespaceAdmin,
	serviceConsumer.ID: serviceConsumer,
	nobody.ID:          nobody,
}
