package runtime

// Kind represents runtime object Kind
type Kind = string

// TypeKind represents type definition of the runtime object, should be embedded into all runtime objects with `yaml:",inline"`
// for proper yaml codec encoding and decoding
type TypeKind struct {
	Kind Kind
}

// GetKind returns Kind
func (tk *TypeKind) GetKind() Kind {
	return tk.Kind
}
