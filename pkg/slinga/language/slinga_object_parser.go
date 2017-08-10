package language

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
)

type SlingaObjectParser struct {
}

func NewSlingaObjectParser() *SlingaObjectParser {
	return &SlingaObjectParser{}
}

func (parser *SlingaObjectParser) parseObject(object *SlingaObject) SlingaObjectInterface {
	switch object.Kind {
	case "service":
		return parser.parseService(object)
	case "context":
		return parser.parseContext(object)
	case "rule":
		return parser.parseRule(object)
	case "cluster":
		return parser.parseCluster(object)
	case "dependency":
		return parser.parseDependency(object)
	case "":
		panic(fmt.Sprintf("Object kind is empty: %v", object))
	default:
		panic(fmt.Sprintf("Unknown object kind: %s (%v)", object.Kind, object))
	}
}

func (parser *SlingaObjectParser) parseService(object *SlingaObject) *Service {
	result := &Service{}
	unmarshalSpec(object, result)
	result.SlingaObject = object
	return result
}

func (parser *SlingaObjectParser) parseContext(object *SlingaObject) *Context {
	result := &Context{}
	unmarshalSpec(object, result)
	result.SlingaObject = object
	return result
}

func (parser *SlingaObjectParser) parseRule(object *SlingaObject) *Rule {
	result := &Rule{}
	unmarshalSpec(object, result)
	result.SlingaObject = object
	return result
}

func (parser *SlingaObjectParser) parseCluster(object *SlingaObject) *Cluster {
	result := &Cluster{}
	unmarshalSpec(object, result)
	result.SlingaObject = object
	return result
}

func (parser *SlingaObjectParser) parseDependency(object *SlingaObject) *Dependency {
	result := &Dependency{}
	unmarshalSpec(object, result)
	result.SlingaObject = object
	return result
}

func unmarshalSpec(object *SlingaObject, result interface{}) {
	e := yaml.DeserializeObject(yaml.SerializeObject(object.Spec), result)
	if e != nil {
		panic(fmt.Sprintf("Unable to unmarshal entry: %s", object.Spec))
	}
}
