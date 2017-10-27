package visualization

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/object"
	"strings"
)

/*
	Escape functions for box/node/edge IDs
*/

func keyEscape(key string) string {
	return strings.NewReplacer("#", "_", ":", "_").Replace(key)
}

/*
	Nodes
*/

type node interface {
	key() string
	render(*diagram)
}

// Internal struct for context node
type contextNode struct {
	contract    *lang.Contract
	contextName string
}

// Returns unique node identifier
func (node contextNode) key() string {
	return keyEscape(strings.Join([]string{object.GetKey(node.contract), node.contextName}, object.KeySeparator))
}

// Renders itself into a graph
func (node contextNode) render(d *diagram) {
	// render namespace box
	namespaceBox{namespace: node.contract.Namespace}.render(d)

	// render contract box
	contractBox{contract: node.contract}.render(d)

	// render context node itself
	addNodeOnce(
		d.graph,
		contractBox{contract: node.contract}.key(), // place into a contract box
		node.key(),
		map[string]string{"label": fmt.Sprintf("Context: %s", node.contextName)},
		d.created,
	)
}

// Internal struct for service instance node
type serviceInstanceNode struct {
	instance *resolve.ComponentInstance
	service  *lang.Service
}

// Returns unique node identifier
func (node serviceInstanceNode) key() string {
	return keyEscape(node.instance.Metadata.Key.GetKey())
}

// Renders itself into a graph
func (node serviceInstanceNode) render(d *diagram) {
	// render namespace box
	namespaceBox{namespace: node.instance.Metadata.Key.Namespace}.render(d)

	// render service box
	serviceBox{service: node.service}.render(d)

	// render context node itself
	addNodeOnce(
		d.graph,
		serviceBox{service: node.service}.key(), // place into a service box
		node.key(),
		map[string]string{"label": node.instance.Metadata.Key.ContextNameWithKeys},
		d.created,
	)
}

// Internal struct for dependency node
type dependencyNode struct {
	dependency *lang.Dependency
}

// Returns unique node identifier
func (node dependencyNode) key() string {
	return keyEscape(object.GetKey(node.dependency))
}

// Renders itself into a graph
func (node dependencyNode) render(d *diagram) {
	// render namespace box
	namespaceBox{namespace: node.dependency.Namespace}.render(d)

	// render dependencies box
	dependencyBox{namespace: node.dependency.Namespace}.render(d)

	// render dependency node itself
	user := d.externalData.UserLoader.LoadUserByID(node.dependency.UserID)
	label := fmt.Sprintf("User: %s (%s)\nContract: %s", user.Name, user.ID, node.dependency.Contract)
	addNodeOnce(
		d.graph,
		dependencyBox{namespace: node.dependency.Namespace}.key(), // place into a dependency box
		node.key(),
		map[string]string{
			"label":     label,
			"style":     "filled",
			"fillcolor": d.getColor(object.GetKey(node.dependency)),
		},
		d.created,
	)
}

// Internal struct for error node
type errorNode struct {
	err error
}

// Returns unique node identifier
func (node errorNode) key() string {
	return keyEscape(strings.Join([]string{"error", fmt.Sprintf("%p", node.err)}, object.KeySeparator)) // nolint: vet
}

// Renders itself into a graph
func (node errorNode) render(d *diagram) {
	// render namespace box
	namespaceBox{namespace: "errors"}.render(d)

	// render error node itself
	addNodeOnce(
		d.graph,
		namespaceBox{namespace: "errors"}.key(), // place into global errors box
		node.key(),
		map[string]string{
			"label":     node.err.Error(),
			"style":     "filled",
			"fillcolor": "firebrick1",
		},
		d.created,
	)
}

/*
	Boxes
*/

// Internal struct for namespace box
type namespaceBox struct {
	namespace string
}

func (box namespaceBox) key() string {
	return keyEscape(strings.Join([]string{"cluster_namespace", box.namespace}, object.KeySeparator))
}

func (box namespaceBox) render(d *diagram) {
	// render namespace box itself
	addSubgraphOnce(
		d.graph,
		"Main", // place into the main graph
		box.key(),
		map[string]string{"label": fmt.Sprintf("NS: %s", box.namespace)},
		d.created,
	)
}

// Internal struct for dependency box
type dependencyBox struct {
	namespace string
}

func (box dependencyBox) key() string {
	return keyEscape(strings.Join([]string{"cluster_namespace", box.namespace, "dependency"}, object.KeySeparator))
}

func (box dependencyBox) render(d *diagram) {
	// render namespace box
	namespaceBox{namespace: box.namespace}.render(d) // nolint: megacheck

	// render dependency box itself
	addSubgraphOnce(
		d.graph,
		namespaceBox{namespace: box.namespace}.key(), // nolint: megacheck
		box.key(),
		map[string]string{"label": "Dependency"},
		d.created,
	)
}

// Internal struct for contract box
type contractBox struct {
	contract *lang.Contract
}

func (box contractBox) key() string {
	return keyEscape(strings.Join([]string{"cluster_contract", object.GetKey(box.contract)}, object.KeySeparator))
}

func (box contractBox) render(d *diagram) {
	// render namespace box
	namespaceBox{namespace: "errors"}.render(d)

	// render contract box itself
	addSubgraphOnce(
		d.graph,
		namespaceBox{namespace: box.contract.Namespace}.key(), // place into a namespace box
		box.key(),
		map[string]string{"label": fmt.Sprintf("Contract: %s", box.contract.Name)},
		d.created,
	)
}

// Internal struct for service box
type serviceBox struct {
	service *lang.Service
}

func (box serviceBox) key() string {
	return keyEscape(strings.Join([]string{"cluster_service", object.GetKey(box.service)}, object.KeySeparator))
}

func (box serviceBox) render(d *diagram) {
	// render namespace box
	namespaceBox{namespace: box.service.Namespace}.render(d)

	// render service box itself
	addSubgraphOnce(
		d.graph,
		namespaceBox{namespace: box.service.Namespace}.key(), // place into a namespace box
		box.key(),
		map[string]string{"label": fmt.Sprintf("Service: %s", box.service.Name)},
		d.created,
	)
}

/*
	Edges
*/

type edge struct {
	src   node
	dst   node
	label string
	color string
}

func (e edge) render(d *diagram) {
	attrs := make(map[string]string)
	if len(e.label) > 0 {
		attrs["label"] = e.label
	}
	if len(e.color) > 0 {
		attrs["color"] = e.color
	}
	addEdge(
		d.graph,
		e.src.key(),
		e.dst.key(),
		attrs,
	)
}
