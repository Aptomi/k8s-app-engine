package runtime

type Displayable interface {
	GetDefaultColumns() []string
	AsColumns() map[string]string
}
