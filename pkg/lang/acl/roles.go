package acl

import "github.com/Aptomi/aptomi/pkg/lang"

var DomainAdmin = &Role{
	ID:   "domain_admin",
	Name: "Domain Admin",
	Privileges: &Privileges{
		Consume: fullAccessAllNS,
		NamespaceObjects: map[string]*Privilege{
			lang.ServiceObject.Kind:    fullAccessAllNS,
			lang.ContractObject.Kind:   fullAccessAllNS,
			lang.DependencyObject.Kind: fullAccessAllNS,
			lang.RuleObject.Kind:       fullAccessAllNS,
		},
		GlobalObjects: map[string]*Privilege{
			lang.ClusterObject.Kind: fullAccessAllNS,
			lang.RuleObject.Kind:    fullAccessAllNS,
		},
	},
}

var NamespaceAdmin = &Role{
	ID:   "namespace_admin",
	Name: "Namespace Admin",
	Privileges: &Privileges{
		Consume: fullAccess,
		NamespaceObjects: map[string]*Privilege{
			lang.ServiceObject.Kind:    fullAccess,
			lang.ContractObject.Kind:   fullAccess,
			lang.DependencyObject.Kind: fullAccess,
			lang.RuleObject.Kind:       fullAccess,
		},
		GlobalObjects: map[string]*Privilege{
			lang.ClusterObject.Kind: viewAccess,
			lang.RuleObject.Kind:    viewAccess,
		},
	},
}

var ServiceConsumer = &Role{
	ID:   "service_consumer",
	Name: "Service Consumer",
	Privileges: &Privileges{
		Consume: fullAccess,
		NamespaceObjects: map[string]*Privilege{
			lang.ServiceObject.Kind:    viewAccess,
			lang.ContractObject.Kind:   viewAccess,
			lang.DependencyObject.Kind: fullAccess,
			lang.RuleObject.Kind:       viewAccess,
		},
		GlobalObjects: map[string]*Privilege{
			lang.ClusterObject.Kind: viewAccess,
			lang.RuleObject.Kind:    viewAccess,
		},
	},
}

var Nobody = &Role{
	ID:         "nobody",
	Name:       "Nobody",
	Privileges: &Privileges{},
}

var Roles = map[string]*Role{
	DomainAdmin.ID:     DomainAdmin,
	NamespaceAdmin.ID:  NamespaceAdmin,
	ServiceConsumer.ID: ServiceConsumer,
}
