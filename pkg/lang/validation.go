package lang

import (
	"context"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/util"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
	"strings"
)

func makePolicyValidator() *validator.Validate {
	result := validator.New()

	// independent validators
	_ = result.RegisterValidation("identifier", validateIdentifier)
	_ = result.RegisterValidation("clustertype", validateClusterType)
	_ = result.RegisterValidation("codetype", validateCodeType)
	_ = result.RegisterValidation("expression", validateExpression)
	_ = result.RegisterValidation("template", validateTemplate)
	_ = result.RegisterValidation("templateNestedMap", validateTemplateNestedMap)
	_ = result.RegisterValidation("labels", validateLabels)
	_ = result.RegisterValidation("labelOperations", validateLabelOperations)
	_ = result.RegisterValidation("allowReject", validateAllowRejectAction)
	_ = result.RegisterValidation("addRoleNS", validateAclRoleActionMap)

	// validators with context containing policy
	result.RegisterStructValidation(validateRule, Rule{})
	result.RegisterStructValidationCtx(validateService, Service{})
	result.RegisterStructValidationCtx(validateDependency, Dependency{})
	result.RegisterStructValidationCtx(validateContract, Contract{})

	return result
}

// checks if a given string is a valid allow/reject action type
func validateAllowRejectAction(fl validator.FieldLevel) bool {
	return util.ContainsString([]string{"allow", "reject"}, fl.Field().String())
}

// checks if a given string is a valid cluster type
func validateClusterType(fl validator.FieldLevel) bool {
	return util.ContainsString([]string{"kubernetes"}, fl.Field().String())
}

// checks if a given string is a valid code type
func validateCodeType(fl validator.FieldLevel) bool {
	return util.ContainsString([]string{"helm", "aptomi/code/kubernetes-helm"}, fl.Field().String())
}

// checks if a given string (or a list of strings) is valid identifier(s)
func validateIdentifier(fl validator.FieldLevel) bool {
	return isIdentifier(fl.Field().String())
}

// checks if a given string (or a list of strings) is valid expression(s)
func validateExpression(fl validator.FieldLevel) bool {
	expr, err := expression.NewExpression(fl.Field().String())
	return expr != nil && err == nil
}

// checks if a given string (or a list of strings) is valid template(s)
func validateTemplate(fl validator.FieldLevel) bool {
	tmpl, err := template.NewTemplate(fl.Field().String())
	return tmpl != nil && err == nil
}

// / checks if a given nested map is a valid map of text templates (e.g. code parameters, discovery parameters, etc)
func validateTemplateNestedMap(fl validator.FieldLevel) bool {
	pMap := fl.Field().Interface().(util.NestedParameterMap)
	_, err := util.ProcessParameterTree(pMap, nil, nil, util.ModeCompile)
	return err == nil
}

// checks if a given map is a valid map of label operations (contains only set/remove, and also label names are valid)
func validateLabelOperations(fl validator.FieldLevel) bool {
	ops := fl.Field().Interface().(LabelOperations)
	for opType, operations := range ops {
		if !util.ContainsString([]string{"set", "remove"}, opType) {
			return false
		}
		for name, _ := range operations {
			if !isIdentifier(name) {
				return false
			}
		}
	}
	return true
}

// checks if a given map is a valid map of setting ACL Role actions
func validateAclRoleActionMap(fl validator.FieldLevel) bool {
	addRoleMap := fl.Field().Interface().(map[string]string)
	for roleID, namespaceList := range addRoleMap {
		role := ACLRolesMap[roleID]
		if role == nil {
			return false
		}

		// mark all namespaces for the role
		namespaces := strings.Split(namespaceList, ",")
		for _, namespace := range namespaces {
			if namespace != namespaceAll && !isIdentifier(strings.TrimSpace(namespace)) {
				return false
			}
		}
	}
	return true
}

// checks if a given map[string]string is a valid map of labels
func validateLabels(fl validator.FieldLevel) bool {
	names := fl.Field().MapKeys()
	for _, name := range names {
		if !isIdentifier(name.String()) {
			return false
		}
	}
	return true
}

// checks if service is valid
func validateService(ctx context.Context, sl validator.StructLevel) {
	service := sl.Current().Interface().(Service)

	// service should have either code or contract set in its components
	policy := ctx.Value("policy").(*Policy)
	for _, component := range service.Components {
		cnt := 0
		if component.Code != nil {
			cnt++
		}
		if len(component.Contract) > 0 {
			cnt++
		}
		if cnt != 1 {
			sl.ReportError(service, fmt.Sprintf("Component[%s].Code|Contract", component.Name), "", "single", "")
			return
		}

		// if contract is set, it should point to an existing contract
		if len(component.Contract) > 0 {
			obj, err := policy.GetObject(ContractObject.Kind, component.Contract, service.Namespace)
			if obj == nil || err != nil {
				sl.ReportError(service, fmt.Sprintf("Component[%s].Contract[%s]", component.Name, component.Contract), "", "exists", "")
				return
			}
		}
	}

	// components should not have duplicate names
	componentNames := make(map[string]bool)
	for _, component := range service.Components {
		if _, exists := componentNames[component.Name]; exists {
			sl.ReportError(service, fmt.Sprintf("Component[%s].Name", component.Name), "", "unique", "")
			return
		}
		componentNames[component.Name] = true
	}

	// components should not have cycles
	_, err := service.GetComponentsSortedTopologically()
	if err != nil {
		sl.ReportError(service, "Components", "", "noCycle", "")
	}

	// dependencies should point to existing components
	for _, component := range service.Components {
		for _, dependencyName := range component.Dependencies {
			if _, exists := componentNames[dependencyName]; !exists {
				sl.ReportError(service, fmt.Sprintf("Component[%s].Dependencies[%s]", component.Name, dependencyName), "", "valid", "")
				return
			}
		}
	}
}

// checks if dependency is valid
func validateDependency(ctx context.Context, sl validator.StructLevel) {
	dependency := sl.Current().Interface().(Dependency)
	policy := ctx.Value("policy").(*Policy)

	// dependency should point to an existing contract
	obj, err := policy.GetObject(ContractObject.Kind, dependency.Contract, dependency.Namespace)
	if obj == nil || err != nil {
		sl.ReportError(dependency, fmt.Sprintf("Contract[%s]", dependency.Contract), "", "exists", "")
		return
	}
}

// checks if contract is valid
func validateContract(ctx context.Context, sl validator.StructLevel) {
	contract := sl.Current().Interface().(Contract)
	policy := ctx.Value("policy").(*Policy)

	// every context should point to an existing service
	for _, ctx := range contract.Contexts {
		serviceName := ""
		if ctx.Allocation != nil {
			serviceName = ctx.Allocation.Service
		}
		obj, err := policy.GetObject(ServiceObject.Kind, serviceName, contract.Namespace)
		if obj == nil || err != nil {
			sl.ReportError(contract, fmt.Sprintf("Contexts[%s].Service[%s]", ctx.Name, serviceName), "", "exists", "")
			return
		}
	}
}

// checks if rule is valid
func validateRule(sl validator.StructLevel) {
	rule := sl.Current().Interface().(Rule)

	// regular rule should have at least one of the actions set
	if rule.Metadata.Kind == RuleObject.Kind {
		hasActions := false
		hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.ChangeLabels) > 0)
		hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.Dependency) > 0)
		hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.Ingress) > 0)
		if !hasActions {
			sl.ReportError(rule.Actions, "Actions", "", "required", "")
		}
		return
	}

	// ACL rule should have its action set
	if rule.Metadata.Kind == ACLRuleObject.Kind {
		hasActions := false
		hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.AddRole) > 0)
		if !hasActions {
			sl.ReportError(rule.Actions, "Actions.AddRole", "", "required", "")
			return
		}
		return
	}
}

func isIdentifier(id string) bool {
	ok, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_-]{0,63}$", id)
	return ok && err == nil
}
