package engine

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/template"
)

// This method defines which contextual information will be exposed to the template engine (for evaluating all templates - discovery, code params, etc)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForAllocationTemplate() *template.TemplateParameters {
	return template.NewTemplateParams(
		struct {
			User      interface{}
			Labels    interface{}
		}{
			User:      node.proxyUser(node.user),
			Labels:    node.labels.Labels,
		},
	)
}

// This method defines which contextual information will be exposed to the template engine (for evaluating all templates - discovery, code params, etc)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForCodeDiscoveryTemplate() *template.TemplateParameters {
	return template.NewTemplateParams(
		struct {
			User      interface{}
			Labels    interface{}
			Discovery interface{}
		}{
			User:      node.proxyUser(node.user),
			Labels:    node.componentLabels.Labels,
			Discovery: node.proxyDiscovery(node.discoveryTreeNode, node.componentKey),
		},
	)
}
