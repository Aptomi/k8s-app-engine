package lang

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/lang/template"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
	english "github.com/go-playground/locales/en"
	"github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/go-playground/validator.v9/translations/en"
)

// Constants
var (
	identifierRegex = "^[a-zA-Z][a-zA-Z0-9_-]{0,63}$"
	clusterTypes    = []string{"kubernetes"}
	codeTypes       = []string{"helm", "raw"}
	labelOpsKeys    = []string{"set", "remove"}
	allowReject     = []string{"allow", "reject"}
)

// Custom type for context key, so we don't have to use 'string' directly
type contextKey string

var policyKey = contextKey("policy")
var errorsKey = contextKey("errors")

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
	result.RegisterValidationCtx("identifier", validateIdentifier)               // nolint: errcheck
	result.RegisterValidationCtx("clustertype", validateClusterType)             // nolint: errcheck
	result.RegisterValidationCtx("codetype", validateCodeType)                   // nolint: errcheck
	result.RegisterValidationCtx("expression", validateExpression)               // nolint: errcheck
	result.RegisterValidationCtx("template", validateTemplate)                   // nolint: errcheck
	result.RegisterValidationCtx("templateNestedMap", validateTemplateNestedMap) // nolint: errcheck
	result.RegisterValidationCtx("labels", validateLabels)                       // nolint: errcheck
	result.RegisterValidationCtx("labelOperations", validateLabelOperations)     // nolint: errcheck
	result.RegisterValidationCtx("allowReject", validateAllowRejectAction)       // nolint: errcheck
	result.RegisterValidationCtx("addRoleNS", validateACLRoleActionMap)          // nolint: errcheck

	// validators with context containing policy
	result.RegisterStructValidation(validateRule, Rule{})
	result.RegisterStructValidation(validateACLRule, ACLRule{})
	result.RegisterStructValidation(validateCluster, Cluster{})
	result.RegisterStructValidationCtx(validateBundle, Bundle{})
	result.RegisterStructValidationCtx(validateClaim, Claim{})
	result.RegisterStructValidationCtx(validateService, Service{})

	// context
	ctx := context.WithValue(context.Background(), policyKey, policy)
	ctx = context.WithValue(ctx, errorsKey, &policyValidationError{})

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
		{
			tag:         "clustertype",
			translation: fmt.Sprintf("'{0}' is not valid, must be in %s", clusterTypes),
		},
		{
			tag:         "codetype",
			translation: fmt.Sprintf("'{0}' is not valid, must be in %s", codeTypes),
		},
		{
			tag:         "allowReject",
			translation: fmt.Sprintf("'{0}' is not valid, must be in %s", allowReject),
		},
		{
			tag:         "systemNS",
			translation: fmt.Sprintf("'{0}' is not valid, must always be '%s'", runtime.SystemNS),
		},
		{
			tag:         "identifier",
			translation: fmt.Sprintf("'{0}' is not a valid identifier"),
		},
		{
			tag:         "expression",
			translation: fmt.Sprintf("'{0}' is not a valid expression"),
		},
		{
			tag:         "template",
			translation: fmt.Sprintf("'{0}' is not a valid text template"),
		},
		{
			tag:         "templateNestedMap",
			translation: fmt.Sprintf("is not a valid text template map (one or more nested text templates is invalid)"),
		},
		{
			tag:         "labels",
			translation: fmt.Sprintf("is not a valid label map (one or more nested label names is invalid)"),
		},
		{
			tag:         "labelOperations",
			translation: fmt.Sprintf("is not a valid label operations map (keys must be in %s, all label names must be valid)", labelOpsKeys),
		},
		{
			tag:         "addRoleNS",
			translation: fmt.Sprintf("is not a valid role assignment map (key must be in %s, namespace list must be comma-separated identifiers/wildcards)", util.GetSortedStringKeys(ACLRolesMap)),
		},
		{
			tag:         "exists",
			translation: fmt.Sprintf("object '{0}' does not exist"),
		},
		{
			tag:         "codeServiceSingle",
			translation: fmt.Sprintf("component '{0}' should either be code or service"),
		},
		{
			tag:         "unique",
			translation: fmt.Sprintf("'{0}' is not unique"),
		},
		{
			tag:         "topologicalSort",
			translation: fmt.Sprintf("{0}"),
		},
		{
			tag:         "ruleActions",
			translation: fmt.Sprintf("is a required field (at least one action must be specified)"),
		},
		{
			tag:         "aclRuleActions",
			translation: fmt.Sprintf("is a required field (role assignment map must be specified)"),
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
	t, err := ut.T(fe.Tag(), reflect.ValueOf(fe.Value()).String(), fe.Param())
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
	vErrors := err.(validator.ValidationErrors) // nolint: errcheck
	for _, vErr := range vErrors {
		errStr := fmt.Sprintf("%s: %s", vErr.Namespace(), vErr.Translate(v.trans))
		result.addError(errStr)
	}

	// collect additional errors stored in context
	for _, errStr := range v.ctx.Value(errorsKey).(*policyValidationError).errList { // nolint: errcheck
		result.addError(errStr)
	}

	return result
}

// adds validation error to the context
func attachErrorToContext(ctx context.Context, level validator.FieldLevel, errMsg string) {
	pve := ctx.Value(errorsKey).(*policyValidationError) // nolint: errcheck
	pve.errList = append(pve.errList, errMsg)
}

// checks in a given field is a string, and it has a valid value (one of the values from a given string array)
func validateInStringArray(ctx context.Context, expectedValues []string, fl validator.FieldLevel) bool {
	return util.ContainsString(expectedValues, fl.Field().String())
}

// checks if a given string is a valid allow/reject action type
func validateAllowRejectAction(ctx context.Context, fl validator.FieldLevel) bool {
	return validateInStringArray(ctx, allowReject, fl)
}

// checks if a given string is a valid cluster type
func validateClusterType(ctx context.Context, fl validator.FieldLevel) bool {
	return validateInStringArray(ctx, clusterTypes, fl)
}

// checks if a given string is a valid code type
func validateCodeType(ctx context.Context, fl validator.FieldLevel) bool {
	return validateInStringArray(ctx, codeTypes, fl)
}

// checks if a given string is valid identifier
func validateIdentifier(ctx context.Context, fl validator.FieldLevel) bool {
	return isIdentifier(fl.Field().String())
}

// checks if a given string is valid expression
func validateExpression(ctx context.Context, fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, err := expression.NewExpression(value)
	if err != nil {
		attachErrorToContext(ctx, fl, err.Error())
	}
	return err == nil
}

// checks if a given string is valid template
func validateTemplate(ctx context.Context, fl validator.FieldLevel) bool {
	_, err := template.NewTemplate(fl.Field().String())
	if err != nil {
		attachErrorToContext(ctx, fl, err.Error())
	}
	return err == nil
}

// checks if a given nested map is a valid map of text templates (e.g. code parameters, discovery parameters, etc)
func validateTemplateNestedMap(ctx context.Context, fl validator.FieldLevel) bool {
	pMap := fl.Field().Interface().(util.NestedParameterMap) // nolint: errcheck
	_, err := util.ProcessParameterTree(pMap, nil, nil, util.ModeCompile)
	if err != nil {
		attachErrorToContext(ctx, fl, err.Error())
	}
	return err == nil
}

// checks if a given map is a valid map of label operations (contains only set/remove, and also label names are valid)
func validateLabelOperations(ctx context.Context, fl validator.FieldLevel) bool {
	ops := fl.Field().Interface().(LabelOperations) // nolint: errcheck
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
func validateACLRoleActionMap(ctx context.Context, fl validator.FieldLevel) bool {
	addRoleMap := fl.Field().Interface().(map[string]string) // nolint: errcheck
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
func validateLabels(ctx context.Context, fl validator.FieldLevel) bool {
	names := fl.Field().MapKeys()
	for _, name := range names {
		if !isIdentifier(name.String()) {
			return false
		}
	}
	return true
}

// checks if bundle is valid
func validateBundle(ctx context.Context, sl validator.StructLevel) {
	bundle := sl.Current().Addr().Interface().(*Bundle) // nolint: errcheck

	// bundle should have either code or service set in its components
	policy := ctx.Value(policyKey).(*Policy) // nolint: errcheck
	for _, component := range bundle.Components {
		cnt := 0
		if component.Code != nil {
			cnt++
		}
		if len(component.Service) > 0 {
			cnt++
		}
		if cnt != 1 {
			sl.ReportError(component.Name, fmt.Sprintf("Component[%s]", component.Name), "", "codeServiceSingle", "")
			return
		}

		// if service is set, it should point to an existing service
		if len(component.Service) > 0 {
			obj, err := policy.GetObject(TypeService.Kind, component.Service, bundle.Namespace)
			if obj == nil || err != nil {
				sl.ReportError(component.Service, fmt.Sprintf("Component[%s].Service[%s/%s]", component.Name, bundle.Namespace, component.Service), "", "exists", "")
				return
			}
		}
	}

	// components should not have duplicate names
	componentNames := make(map[string]bool)
	for _, component := range bundle.Components {
		if _, exists := componentNames[component.Name]; exists {
			sl.ReportError(component.Name, fmt.Sprintf("Component[%s].Name", component.Name), "", "unique", "")
			return
		}
		componentNames[component.Name] = true
	}

	// components should not have cycles
	_, err := bundle.GetComponentsSortedTopologically()
	if err != nil {
		sl.ReportError(err.Error(), "Components", "", "topologicalSort", "")
	}

	// dependencies between bundle components should be valid
	for _, component := range bundle.Components {
		for _, componentName := range component.Dependencies {
			if _, exists := componentNames[componentName]; !exists {
				sl.ReportError(componentName, fmt.Sprintf("Component[%s].Dependencies[%s]", component.Name, componentName), "", "exists", "")
				return
			}
		}
	}
}

// checks if claim is valid
func validateClaim(ctx context.Context, sl validator.StructLevel) {
	claim := sl.Current().Addr().Interface().(*Claim) // nolint: errcheck
	policy := ctx.Value(policyKey).(*Policy)          // nolint: errcheck

	// claim should point to an existing service
	obj, err := policy.GetObject(TypeService.Kind, claim.Service, claim.Namespace)
	if obj == nil || err != nil {
		sl.ReportError(claim.Service, fmt.Sprintf("Service[%s/%s]", claim.Namespace, claim.Service), "", "exists", "")
		return
	}
}

// checks if service is valid
func validateService(ctx context.Context, sl validator.StructLevel) {
	service := sl.Current().Addr().Interface().(*Service) // nolint: errcheck
	policy := ctx.Value(policyKey).(*Policy)              // nolint: errcheck

	// every context should point to an existing bundle
	for _, serviceCtx := range service.Contexts {
		bundleName := ""
		if serviceCtx.Allocation != nil {
			bundleName = serviceCtx.Allocation.Bundle
		}
		obj, err := policy.GetObject(TypeBundle.Kind, bundleName, service.Namespace)
		if obj == nil || err != nil {
			sl.ReportError(bundleName, fmt.Sprintf("Contexts[%s].Bundle[%s/%s]", serviceCtx.Name, service.Namespace, bundleName), "", "exists", "")
			return
		}
	}
}

// checks if rule is valid
func validateRule(sl validator.StructLevel) {
	rule := sl.Current().Addr().Interface().(*Rule) // nolint: errcheck

	// rule should have at least one of the actions set
	hasActions := false
	hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.ChangeLabels) > 0)
	hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.Claim) > 0)
	hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.Ingress) > 0)
	if !hasActions {
		sl.ReportError(rule.Actions, "Actions", "", "ruleActions", "")
		return
	}
}

// checks if ACL rule is valid
func validateACLRule(sl validator.StructLevel) {
	rule := sl.Current().Addr().Interface().(*ACLRule) // nolint: errcheck

	// ACL rule should have its action set
	hasActions := false
	hasActions = hasActions || (rule.Actions != nil && len(rule.Actions.AddRole) > 0)
	if !hasActions {
		sl.ReportError(rule.Actions, "Actions.AddRole", "", "aclRuleActions", "")
		return
	}
}

// checks if cluster is valid
func validateCluster(sl validator.StructLevel) {
	cluster := sl.Current().Addr().Interface().(*Cluster) // nolint: errcheck
	if cluster.Namespace != runtime.SystemNS {
		sl.ReportError(cluster.Namespace, "Namespace", "", "systemNS", "")
	}
}

func isIdentifier(id string) bool {
	ok, err := regexp.MatchString(identifierRegex, id)
	return ok && err == nil
}
