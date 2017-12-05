package runtime

// Displayable represents object that could be represented as columns and have some default set of columns to be shown
type Displayable interface {
	GetDefaultColumns() []string
	AsColumns() map[string]string
}
