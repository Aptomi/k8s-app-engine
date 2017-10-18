package lang

import (
	"github.com/Aptomi/aptomi/pkg/object"
)

// ACLRulesBootstrap is a set of default ACL rules, which Aptomi will get initialized with on the first run
var ACLRulesBootstrap = []*ACLRule{
	// domain admins
	{
		Metadata: Metadata{
			Kind:      ACLRuleObject.Kind,
			Namespace: object.SystemNS,
			Name:      "aptomi_bootstrap_" + domainAdmin.ID,
		},
		Weight:   100,
		Criteria: &Criteria{RequireAll: []string{"role == 'aptomi_domain_admin'"}},
		Actions: &RuleActions{
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(labelRole, domainAdmin.ID)),
		},
	},
	// namespace admins for 'main' namespace
	{
		Metadata: Metadata{
			Kind:      ACLRuleObject.Kind,
			Namespace: object.SystemNS,
			Name:      "aptomi_bootstrap_" + namespaceAdmin.ID,
		},
		Weight:   200,
		Criteria: &Criteria{RequireAll: []string{"role == 'aptomi_main_ns_admin'"}},
		Actions: &RuleActions{
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(labelRole, namespaceAdmin.ID)),
			Namespaces:   map[string]string{"main": "true"},
		},
	},
	// service consumers for 'main' namespace
	{
		Metadata: Metadata{
			Kind:      ACLRuleObject.Kind,
			Namespace: object.SystemNS,
			Name:      "aptomi_bootstrap_" + serviceConsumer.ID,
		},
		Weight:   300,
		Criteria: &Criteria{RequireAll: []string{"role == 'aptomi_main_ns_consumer'"}},
		Actions: &RuleActions{
			ChangeLabels: ChangeLabelsAction(NewLabelOperationsSetSingleLabel(labelRole, serviceConsumer.ID)),
			Namespaces:   map[string]string{"main": "true"},
		},
	},
}
