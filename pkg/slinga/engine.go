package slinga

import (
	"bytes"
	"errors"
	"strings"
	"text/template"
)

/*
	Core engine for Slinga processing and evaluation
*/

// TemplateData defines struct that holds all params and gets passed to template engine for evaluating template string
type TemplateData struct {
	Labels    map[string]string
	User      User
	Discovery map[string]interface{}
}

// ResolveUsage evaluates all recorded Dependencies ("<user> needs <service> with <labels>") and calculates allocations
func (usage *ServiceUsageState) ResolveUsage(users *GlobalUsers) error {
	for serviceName, dependencies := range usage.Dependencies.Dependencies {
		for _, d := range dependencies {
			user := users.Users[d.UserID]

			// take user labels
			labels := user.getLabelSet()

			// combine them with dependency labels
			labels = labels.addLabels(d.getLabelSet())

			// see if it needs to be traced (addl debug output on console)
			tracing.setEnable(d.Trace)

			// resolve usage via applying policy
			resKey, err := usage.resolveWithLabels(user, serviceName, labels, usage.ComponentInstanceMap, 0)

			// disable tracing
			tracing.setEnable(false)

			if err != nil {
				return err
			}
			d.ResolvesTo = resKey
		}
	}
	return nil
}

// Evaluate "<user> needs <service>" statement
func (usage *ServiceUsageState) resolveWithLabels(user User, serviceName string, labels LabelSet, cim map[string]interface{}, depth int) (string, error) {

	// Resolving allocations for service
	debug.Infof("Resolving allocations for service %s (user = %s, labels = %s)", serviceName, user.Name, labels)

	tracing.Printf(depth, "[Dependency]")
	tracing.Printf(depth, "User: %s (ID = %s)", user.Name, user.ID)
	tracing.Printf(depth+1, "Labels: %s", labels)
	tracing.Printf(depth, "Service: %s", serviceName)

	// Policy
	policy := usage.Policy

	// Locate the service
	service, err := policy.getService(serviceName)
	if err != nil {
		tracing.Printf(depth, "Error while trying to look up service %s (%v)", serviceName, err)
		return "", err
	}

	// Process service and transform labels
	labels = labels.applyTransform(service.Labels)
	tracing.Printf(depth+1, "New labels = %s", labels)

	// Match the context
	context, err := policy.getMatchedContext(*service, user, labels, depth)
	if err != nil {
		tracing.Printf(depth, "Error while matching context for service %s (%v)", serviceName, err)
		return "", err
	}
	// If no matching context is found, let's just exit
	if context == nil {
		tracing.Printf(depth, "No context matched for service %s", serviceName)
		return "", nil
	}

	// Process context and transform labels
	labels = labels.applyTransform(context.Labels)
	tracing.Printf(depth, "Context: %s", context.Name)
	tracing.Printf(depth+1, "New labels = %s", labels)

	// Match the allocation
	allocation, err := policy.getMatchedAllocation(*service, user, *context, labels, depth)
	if err != nil {
		tracing.Printf(depth, "Error while matching allocation for service %s, context %s (%v)", serviceName, context.Name, err)
		return "", err
	}
	// If no matching allocation is found, let's just exit
	if allocation == nil {
		tracing.Printf(depth, "No allocation matched for service %s, context %s", serviceName, context.Name)
		return "", nil
	}

	// Process allocation and transform labels
	labels = labels.applyTransform(allocation.Labels)
	tracing.Printf(depth, "Allocation: %s", allocation.NameResolved)
	tracing.Printf(depth+1, "New labels = %s", labels)

	// Now, sort all components in topological order
	componentsOrdered, err := service.getComponentsSortedTopologically()
	if err != nil {
		return "", err
	}

	// Resolve every component
	for _, component := range componentsOrdered {
		// Process component and transform labels
		componentLabels := labels.applyTransform(component.Labels)

		// Is it a code?
		var codeParams interface{}
		var discoveryParams interface{}

		componentKey := usage.createServiceUsageKey(service, context, allocation, component)
		discoveryParams, err = component.processTemplateParams(component.Discovery, componentKey, componentLabels, user, cim, "discovery", depth)
		if err != nil {
			return "", err
		}

		// Create new map with resolution keys for component
		cimComponent := make(map[string]interface{})
		cim[component.Name] = cimComponent

		if component.Code != nil {
			// Evaluate code params
			debug.Infof("Processing dependency on code execution: %s (in %s)", component.Name, service.Name)

			// TODO: WAT?? no HelmName please
			cimComponent["instance"] = HelmName(componentKey)
			if discoveryParamsMap, ok := discoveryParams.(map[interface{}]interface{}); ok {
				for k, v := range discoveryParamsMap {
					cimComponent[k.(string)] = v
				}
			}

			codeParams, err = component.processTemplateParams(component.Code.Params, componentKey, componentLabels, user, cim, "code", depth)
			if err != nil {
				return "", err
			}
		} else if component.Service != "" {
			debug.Infof("Processing dependency on another service: %s -> %s (in %s)", component.Name, component.Service, service.Name)
			tracing.Println()

			_, err := usage.resolveWithLabels(user, component.Service, componentLabels, cimComponent, depth+1)
			if err != nil {
				return "", err
			}
		} else {
			debug.Fatalf("Invalid component: %s (in %s)", component.Name, service.Name)
		}

		// Record usage of a given component
		usage.recordUsage(componentKey, user, componentLabels, codeParams, discoveryParams)
	}

	// Record usage of a given service
	serviceKey := usage.createServiceUsageKey(service, context, allocation, nil)
	usage.recordUsage(serviceKey, user, labels, nil, nil)

	return context.Name + "#" + allocation.NameResolved, nil
}

// Topologically sort components and return true if there is a cycle detected
func (service *Service) dfsComponentSort(u *ServiceComponent, colors map[string]int) bool {
	colors[u.Name] = 1

	for _, vName := range u.Dependencies {
		v, exists := service.getComponentsMap()[vName]
		if !exists {
			debug.Fatalf("Invalid dependency in service %s: %s", service.Name, vName)
		}
		if vColor, ok := colors[v.Name]; !ok {
			// not visited yet -> visit and exit if a cycle was found
			if service.dfsComponentSort(v, colors) {
				return true
			}
		} else if vColor == 1 {
			return true
		}
	}

	service.componentsOrdered = append(service.componentsOrdered, u)
	colors[u.Name] = 2
	return false
}

// Sorts all components in a topological way
func (service *Service) getComponentsSortedTopologically() ([]*ServiceComponent, error) {
	if service.componentsOrdered == nil {
		// Initiate colors
		colors := make(map[string]int)

		// Dfs
		var cycle = false
		for _, c := range service.Components {
			if _, ok := colors[c.Name]; !ok {
				if service.dfsComponentSort(c, colors) {
					cycle = true
					break
				}
			}
		}

		if cycle {
			return nil, errors.New("Component cycle detected in service " + service.Name)
		}
	}

	return service.componentsOrdered, nil
}

// Helper to get a service
func (policy *Policy) getService(serviceName string) (*Service, error) {
	// Locate the service
	service := policy.Services[serviceName]
	if service == nil {
		return nil, errors.New("Service " + serviceName + " not found")
	}
	return service, nil
}

// Helper to get a matched context
func (policy *Policy) getMatchedContext(service Service, user User, labels LabelSet, depth int) (*Context, error) {
	// Locate the list of contexts for service
	contexts, ok := policy.Contexts[service.Name]
	if !ok || len(contexts) <= 0 {
		tracing.Printf(depth+1, "No contexts found")
		return nil, errors.New("No contexts found for " + service.Name)
	}

	// See which context matches
	var contextMatched *Context
	for _, c := range contexts {
		m := c.matches(labels)
		tracing.Printf(depth+1, "[%t] Testing context '%s': (criteria = %+v)", m, c.Name, c.Criteria)
		if m {
			contextMatched = c
			break
		}
	}

	if contextMatched != nil {
		debug.Infof("Matched context: '%s' (service = %s, user = %s)", contextMatched.Name, service.Name, user.Name)
	} else {
		debug.Infof("No context matched (service = %s, user = %s)", service.Name, user.Name)
	}
	return contextMatched, nil
}

// Helper to get a matched allocation
func (policy *Policy) getMatchedAllocation(service Service, user User, context Context, labels LabelSet, depth int) (*Allocation, error) {
	if len(context.Allocations) <= 0 {
		tracing.Printf(depth+1, "No allocations found")
		return nil, errors.New("No allocations found for " + service.Name)
	}

	// See which allocation matches
	var allocationMatched *Allocation
	for _, a := range context.Allocations {
		m := a.matches(labels)
		tracing.Printf(depth+1, "[%t] Testing allocation '%s': (criteria = %+v)", m, a.Name, a.Criteria)
		if m {
			allocationMatched = a
			break
		}
	}

	// Check errors and resolve allocation name (it can be dynamic, depending on user labels)
	if allocationMatched != nil {
		err := allocationMatched.resolveName(user)
		if err != nil {
			debug.Infof("Cannot resolve name for an allocation: '%s' (context = %s, service = %s, user = %s)", allocationMatched.Name, context.Name, service.Name, user.Name)
			return nil, nil
		}
		debug.Infof("Matched allocation: '%s' -> '%s' (context = %s, service = %s, user = %s)", allocationMatched.Name, allocationMatched.NameResolved, context.Name, service.Name, user.Name)
	} else {
		debug.Infof("No allocation matched (context = %s, service = %s, user = %s)", context.Name, service.Name, user.Name)
	}

	return allocationMatched, nil
}

type ProcessingError struct {
	Reason string
}

func (err ProcessingError) Error() string {
	return err.Reason
}

func (component *ServiceComponent) processTemplateParams(template interface{}, componentKey string, labels LabelSet, user User, cim map[string]interface{}, templateType string, depth int) (interface{}, error) {
	if template == nil {
		return nil, nil
	}
	tracing.Printf(depth+1, "Component: %s (%s)", component.Name, templateType)

	cimCopy := make(map[string]interface{})
	// TODO SPARTAAA!
	cimCopy["instance"] = HelmName(componentKey)
	for k, v := range cim {
		cimCopy[k] = v
	}

	templateData := TemplateData{
		Labels:    labels.Labels,
		Discovery: cimCopy,
		User:      user}

	var evalParamsInterface func(params interface{}) (interface{}, error)
	evalParamsInterface = func(params interface{}) (interface{}, error) {
		if params == nil {
			return "", nil
		} else if paramsMap, ok := params.(map[interface{}]interface{}); ok {
			resultMap := make(map[interface{}]interface{})

			for key, value := range paramsMap {
				evaluatedValue, err := evalParamsInterface(value)
				if err != nil {
					return nil, err
				}
				resultMap[key] = evaluatedValue
			}

			return resultMap, nil
		} else if paramsStr, ok := params.(string); ok {
			evaluatedValue, err := evaluateCodeParamTemplate(paramsStr, templateData)
			tracing.Printf(depth+2, "Parameter '%s': %s", paramsStr, evaluatedValue)
			if err != nil {
				return nil, err
			}
			return evaluatedValue, nil
		} else if paramsInt, ok := params.(int); ok {
			return paramsInt, nil
		} else if paramsBool, ok := params.(bool); ok {
			return paramsBool, nil
		}

		return nil, ProcessingError{"There should be map[string]interface{} or string"}
	}

	return evalParamsInterface(template)
}

func evaluateCodeParamTemplate(templateStr string, templateData TemplateData) (string, error) {
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		return "", errors.New("Invalid template " + templateStr)
	}

	var doc bytes.Buffer
	err = tmpl.Execute(&doc, templateData)

	if err != nil {
		return "", errors.New("Cannot evaluate template " + templateStr)
	}

	result := doc.String()
	if strings.Contains(result, "<no value>") {
		return "", errors.New("Cannot evaluate template " + templateStr)
	}

	// TODO: remove this ugly shit later (should not reference Helm in generic engine)
	return HelmName(doc.String()), nil
}
