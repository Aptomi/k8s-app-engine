package acl

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
)

const NamespaceAll = "*"
const LabelRole = "role"

type Role struct {
	ID         string
	Name       string
	Privileges *Privileges
}

type Privileges struct {
	// Consume indicate whether or not this role can consume services
	Consume *Privilege

	// NamespaceObjects specifies whether or not this role can view/manage a certain object kind within a non-system namespace
	NamespaceObjects map[string]*Privilege

	// GlobalObjects specifies whether or not this role can view/manage a certain object kind within a system namespace
	GlobalObjects map[string]*Privilege
}

type Privilege struct {
	// View indicates whether or not a user can view objects
	View bool

	// Manage indicates whether or not a user can manage objects, i.e. perform CRUD operations
	Manage bool

	// If NamespaceScope equals to NamespaceAll, then user privilege applies to all namespaces.
	// If NamespaceScope equals to a specific namespace, then user privilege applies to this specific namespace only
	// If NamespaceScope is not set, a user privilege applies to a set of given namespaces
	NamespaceScope string
}

// RuleObject is an informational data structure with Kind and Constructor for ACL Rule
var RuleObject = &object.Info{
	Kind:        "aclrule",
	Versioned:   true,
	Constructor: func() object.Base { return &Rule{} },
}

type Rule = lang.Rule

var fullAccessAllNS = &Privilege{
	View:           true,
	Manage:         true,
	NamespaceScope: NamespaceAll,
}

var fullAccess = &Privilege{
	View:   true,
	Manage: true,
}

var viewAccess = &Privilege{
	View: true,
}
