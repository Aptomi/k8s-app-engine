package visibility

type graphNode interface {
	getID() string
	getLabel() string
	getGroup() string
	getEdgeLabel(graphNode) string
}
