package lang

import (
	"context"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/object"
	"github.com/Aptomi/aptomi/pkg/util"
	english "github.com/go-playground/locales/en"
	"github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/go-playground/validator.v9/translations/en"
	"reflect"
	"regexp"
	"strings"
)

// Constants
var (
	identifierRegex = "^[a-zA-Z][a-zA-Z0-9_-]{0,63}$"
	clusterTypes    = []string{"kubernetes"}
	codeTypes       = []string{"helm", "aptomi/code/kubernetes-helm"}
	labelOpsKeys    = []string{"set", "remove"}
	allowReject     = []string{"allow", "reject"}
)

// Custom type for context key, so we don't have to use 'string' directly
type contextKey string

var policyKey = contextKey("policy")

func (c contextKey) String() string {
	return "lang context key " + string(c)
}

// Custom error for policy validation
type policyValidationError struct {
	errList []string
}

func (err policyValidationError) Error() string {
	return strings.Join(err.errList, "\n")
}

func (err *policyValidationError) addError(errStr string) {
	err.errList = append(err.errList, errStr)
}

// PolicyValidator is a custom validator for the policy
type PolicyValidator struct {
	val    *validator.Validate
	ctx    context.Context
	policy *Policy
	trans  ut.Translator
}

// NewPolicyValidator creates a new PolicyValidator
func NewPolicyValidator(policy *Policy) *PolicyValidator {
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
	_ = result.RegisterValidation("addRoleNS", validateACLRoleActionMap)

	// validators with context containing policy
	result.RegisterStructValidation(validateRule, Rule{})
	result.RegisterStructValidation(validateCluster, Cluster{})
	result.RegisterStructValidationCtx(validateService, Service{})
	result.RegisterStructValidationCtx(validateDependency, Dependency{})
	result.RegisterStructValidationCtx(validateContract, Contract{})

	// context
	ctx := context.WithValue(context.Background(), policyKey, policy)

	// default translations
	eng := english.New()
	uni := ut.New(eng, eng)
	trans, _ := uni.GetTranslator("en")
	err := en.RegisterDefaultTranslations(result, trans)
	if err != nil {
		panic(err)
	}

	// additional translations
	translations := []struct {
		tag         string
		translation string
	}{
		// static
		{
			tag:         "identifier",
			translation: fmt.Sprintf("{0} must be a valid identifier, but found '{1}'"),
		},
		{
			tag:         "clustertype",
			translation: fmt.Sprintf("{0} must be in %s, but found '{1}'", clusterTypes),
		},
		{
			tag:         "codetype",
			translation: fmt.Sprintf("{0} must be in %s, but found '{1}'", codeTypes),
		},
		{
			tag:         "expression",
			translation: fmt.Sprintf("{0} must be a valid expression, but found '{1}'"),
		},
		{
			tag:         "template",
			translation: fmt.Sprintf("{0} must be a valid text template, but found '{1}'"),
		},
		{
			tag:         "templateNestedMap",
			translation: fmt.Sprintf("{0} must be a valid text template map (all nested text templates must be valid)"),
		},
		{
			tag:         "labels",
			translation: fmt.Sprintf("{0} must be a valid label map (all label names must be valid)"),
		},
		{
			tag:         "labelOperations",
			translation: fmt.Sprintf("{0} must be a valid label operations map (keys must be in %s, all label names must be valid)", labelOpsKeys),
		},
		{
			tag:         "allowReject",
			translation: fmt.Sprintf("{0} must be in %s, but found '{1}'", allowReject),
		},
		{
			tag:         "addRoleNS",
			translation: fmt.Sprintf("{0} must be a valid role assignment map (key must be in %s, namespace list must be comma-separated identifiers/wildcards)", util.GetSortedStringKeys(ACLRolesMap)),
		},
		// dynamic/custom
		{
			tag:         "exists",
			translation: fmt.Sprintf("object does not exist"),
		},
		{
			tag:         "single",
			translation: fmt.Sprintf("only a single value is allowed"),
		},
		{
			tag:         "unique",
			translation: fmt.Sprintf("must be unique, but it is not"),
		},
		{
			tag:         "noComponentCycle",
			translation: fmt.Sprintf("circular dependency detected in components"),
		},
		{
			tag:         "ruleActions",
			translation: fmt.Sprintf("{0} must have at least one action defined"),
		},
		{
			tag:         "aclRuleActions",
			translation: fmt.Sprintf("{0} is a required field for ACL rule. Must specify role assignment map"),
		},
		{
			tag:         "systemNS",
			translation: fmt.Sprintf("{0} must be '%s', but found '{1}'", object.SystemNS),
		},
	}
	for _, t := range translations {
		err = result.RegisterTranslation(t.tag, trans, registrationFunc(t.tag, t.translation), translateFunc)
		if err != nil {
			panic(err)
		}
	}

	return &PolicyValidator{
		val:    result,
		ctx:    ctx,
		policy: policy,
		trans:  trans,
	}
}

func registrationFunc(tag string, translation string) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) (err error) {
		if err = ut.Add(tag, translation, true); err != nil {
			return
		}
		return
	}
}

func translateFunc(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(fe.Tag(), fe.Field(), reflect.ValueOf(fe.Value()).String())
	if err != nil {
		return fe.(error).Error()
	}
	return t
}

// Validate validates the entire policy for errors and returns an error (it can be casted to
// policyValidationError, containing a list of errors inside). When error is printed as string, it will
// automatically contains the full list of validation errors.
func (v *PolicyValidator) Validate() error {
	// validate policy
	err := v.val.StructCtx(v.ctx, v.policy)
	if err == nil {
		return nil
	}

	// collect human-readable errors
	result := policyValidationError{}
	vErrors := err.(validator.ValidationErrors)
	for _, vErr := range vErrors {
		errStr := fmt.Sprintf("%s: %s", vErr.Namespace(), vErr.Translate(v.trans))
		result.addError(errStr)
	}

	return result
}

// checks if a given string is a valid allow/reject action type
func validateAllowRejectAction(fl validator.FieldLevel) bool {
	return util.ContainsString(allowReject, fl.Field().String())
}

// checks if a given string is a valid cluster type
func validateClusterType(fl validator.FieldLevel) bool {
	return util.ContainsString(clusterTypes, fl.Field().String())
}

// checks if a given string is a valid code type
func validateCodeType(fl validator.FieldLevel) bool {
	return util.ContainsString(codeTypes, fl.Field().String())
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
		if !util.ContainsString(labelOpsKeys, opType) {
			return false
		}
		for name := range operations {
			if !isIdentifier(name) {
				return false
			}
		}
	}
	return true
}

// checks if a given map is a valid map of setting ACL Role actions
func validateACLRoleActionMap(fl validator.FieldLevel) bool {
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
	service := sl.Current().Addr().Interface().(*Service)

	// service should have either code or contract set in its components
	policy := ctx.Value(policyKey).(*Policy)
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
		sl.ReportError(service, "Components", "", "noComponentCycle", "")
	}

	// dependencies should point to existing components
	for _, component := range service.Components {
		for _, dependencyName := range component.Dependencies {
			if _, exists := componentNames[dependencyName]; !exists {
				sl.ReportError(service, fmt.Sprintf("Component[%s].Dependencies[%s]", component.Name, dependencyName), "", "exists", "")
				return
			}
		}
	}
}

// checks if dependency is valid
func validateDependency(ctx context.Context, sl validator.StructLevel) {
	dependency := sl.Current().Addr().Interface().(*Dependency)
	policy := ctx.Value(policyKey).(*Policy)

	// dependency should point to an existing contract
	obj, err := policy.GetObject(ContractObject.Kind, dependency.Contract, dependency.Namespace)
	if obj == nil || err != nil {
		sl.ReportError(dependency, fmt.Sprintf("Contract[%s]", dependency.Contract), "", "exists", "")
		return
	}
}

// checks if contract is valid
func validateContract(ctx context.Context, sl validator.StructLevel) {
	contract := sl.Current().Addr().Interface().(*Contract)
	policy := ctx.Value(policyKey).(*Policy)

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
	rule := sl.Current().Addr().Interface().(*Rule)

	// regular rule should have at least one of the actions set
	if rule.Metadata.Kind == RuleObject.Kind {
		hasActions := false
		hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.ChangeLabels) > 0)
		hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.Dependency) > 0)
		hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.Ingress) > 0)
		if !hasActions {
			sl.ReportError(rule.Actions, "Actions", "", "ruleActions", "")
		}
		return
	}

	// ACL rule should have its action set
	if rule.Metadata.Kind == ACLRuleObject.Kind {
		hasActions := false
		hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.AddRole) > 0)
		if !hasActions {
			sl.ReportError(rule.Actions, "Actions.AddRole", "", "aclRuleActions", "")
			return
		}
		return
	}
}

// checks if cluster is valid
func validateCluster(sl validator.StructLevel) {
	cluster := sl.Current().Addr().Interface().(*Cluster)
	if cluster.Namespace != object.SystemNS {
		sl.ReportError(cluster.Namespace, "Namespace", "", "systemNS", "")
	}
}

func isIdentifier(id string) bool {
	ok, err := regexp.MatchString(identifierRegex, id)
	return ok && err == nil
}
