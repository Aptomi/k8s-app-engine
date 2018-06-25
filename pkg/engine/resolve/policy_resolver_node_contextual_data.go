package resolve

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

/*
	Data exposed to context expressions/criteria defined in policy
*/

// This method defines which contextual information will be exposed to the expression engine (for evaluating criteria)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForContextExpression() *expression.Parameters {
	return expression.NewParams(
		node.labels.Labels,
		map[string]interface{}{},
	)
}

// This method defines which contextual information will be exposed to the expression engine (for evaluating criteria)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForComponentCriteria() *expression.Parameters {
	return expression.NewParams(
		node.labels.Labels,
		map[string]interface{}{},
	)
}

/*
	Data exposed to rules defined
*/

// This method defines which contextual information will be exposed to the expression engine (for evaluating rules)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForRuleExpression() *expression.Parameters {
	return expression.NewParams(
		node.labels.Labels,
		map[string]interface{}{
			"Bundle": node.proxyBundle(node.bundle),
			"Claim":  node.proxyClaim(node.claim),
		},
	)
}

/*
	Data exposed to templates defined in policy
*/

// This method defines which contextual information will be exposed to the template engine (for evaluating all templates - discovery, code params, etc)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForContextAllocationTemplate() *template.Parameters {
	return template.NewParams(
		struct {
			User   interface{}
			Claim  interface{}
			Labels interface{}
		}{
			User:   node.proxyUser(node.user),
			Claim:  node.proxyClaim(node.claim),
			Labels: node.labels.Labels,
		},
	)
}

// This method defines which contextual information will be exposed to the template engine (for evaluating all templates - discovery, code params, etc)
// Be careful about what gets exposed through this method. User can refer to structs and their methods from the policy
func (node *resolutionNode) getContextualDataForCodeDiscoveryTemplate() *template.Parameters {
	return template.NewParams(
		struct {
			User      interface{}
			Labels    interface{}
			Discovery interface{}
			Target    interface{}
		}{
			User:      node.proxyUser(node.user),
			Labels:    node.labels.Labels,
			Discovery: node.proxyDiscovery(node.discoveryTreeNode, node.componentKey),
			Target:    node.proxyTarget(node.componentKey),
		},
	)
}

/*
	Proxy functions
*/

// How bundle is visible from the policy language
func (node *resolutionNode) proxyBundle(bundle *lang.Bundle) interface{} {
	return struct {
		lang.Metadata
		Labels interface{}
	}{
		Metadata: bundle.Metadata,
		Labels:   bundle.Labels,
	}
}

// How user is visible from the policy language
func (node *resolutionNode) proxyUser(user *lang.User) interface{} {
	return struct {
		Name    interface{}
		Labels  interface{}
		Secrets interface{}
	}{
		Name:    user.Name,
		Labels:  user.Labels,
		Secrets: node.resolver.externalData.SecretLoader.LoadSecretsByUserName(user.Name),
	}
}

// How claim is visible from the policy language
func (node *resolutionNode) proxyClaim(claim *lang.Claim) interface{} {
	result := struct {
		lang.Metadata
		ID interface{}
	}{
		Metadata: claim.Metadata,
		ID:       runtime.KeyForStorable(claim),
	}
	return result
}

// How target is visible from the policy language
func (node *resolutionNode) proxyTarget(cik *ComponentInstanceKey) interface{} {
	result := struct {
		Namespace string
	}{
		Namespace: cik.TargetSuffix,
	}
	return result
}

// How discovery tree is visible from the policy language
func (node *resolutionNode) proxyDiscovery(discoveryTree util.NestedParameterMap, cik *ComponentInstanceKey) interface{} {
	result := discoveryTree.MakeCopy()

	// special case to announce own component instance
	result["Instance"] = util.EscapeName(cik.GetDeployName())

	// special case to announce own component ID
	result["InstanceId"] = util.HashFnv(cik.GetKey())

	// expose parent bundle information as well
	if cik.IsComponent() {
		bundleCik := cik.GetParentBundleKey()
		result["Bundle"] = util.NestedParameterMap{
			// announce instance of the enclosing bundle
			"Instance": util.EscapeName(bundleCik.GetDeployName()),
			// announce instance of the enclosing bundle instance ID
			"InstanceId": util.HashFnv(bundleCik.GetKey()),
		}
	}

	return result
}
