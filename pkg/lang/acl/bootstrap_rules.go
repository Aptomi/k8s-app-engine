package acl

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
)

var BootstrapAclRules = []*Rule{
	// every single person is a domain admin
	{
		Metadata: lang.Metadata{
			Kind:      RuleObject.Kind,
			Namespace: object.SystemNS,
			Name:      "aptomi_bootstrap_" + DomainAdmin.ID,
		},
		Weight:   100,
		Criteria: &lang.Criteria{RequireAll: []string{"true"}},
		Actions: &lang.RuleActions{
			ChangeLabels: lang.ChangeLabelsAction(lang.NewLabelOperationsSetSingleLabel(LabelRole, DomainAdmin.ID)),
			Stop:         true,
		},
	},
	// every single person is a namespace admin
	{
		Metadata: lang.Metadata{
			Kind:      RuleObject.Kind,
			Namespace: object.SystemNS,
			Name:      "aptomi_bootstrap_" + NamespaceAdmin.ID,
		},
		Weight:   200,
		Criteria: &lang.Criteria{RequireAll: []string{"true"}},
		Actions: &lang.RuleActions{
			ChangeLabels: lang.ChangeLabelsAction(lang.NewLabelOperationsSetSingleLabel(LabelRole, NamespaceAdmin.ID)),
			Stop:         true,
		},
	},
	// every single person is a service consumer
	{
		Metadata: lang.Metadata{
			Kind:      RuleObject.Kind,
			Namespace: object.SystemNS,
			Name:      "aptomi_bootstrap_" + ServiceConsumer.ID,
		},
		Weight:   300,
		Criteria: &lang.Criteria{RequireAll: []string{"true"}},
		Actions: &lang.RuleActions{
			ChangeLabels: lang.ChangeLabelsAction(lang.NewLabelOperationsSetSingleLabel(LabelRole, ServiceConsumer.ID)),
			Stop:         true,
		},
	},
}
