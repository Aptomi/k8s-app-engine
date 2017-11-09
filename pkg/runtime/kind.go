package runtime

// Kind represents runtime object Kind
type Kind = string

type TypeKind struct {
	Kind Kind
}

func (tk *TypeKind) GetKind() Kind {
	return tk.Kind
}
