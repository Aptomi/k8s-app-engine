package visualization

type edge struct {
	src   graphNode
	dst   graphNode
	label string
}

func newEdge(src graphNode, dst graphNode, label string) *edge {
	return &edge{src: src, dst: dst, label: label}
}

func (e *edge) getSrc() graphNode {
	return e.src
}

func (e *edge) getDst() graphNode {
	return e.dst
}

func (e *edge) getLabel() string {
	return e.label
}
