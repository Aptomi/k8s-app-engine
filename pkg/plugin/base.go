package plugin

// Plugin is a base interface for all engine plugins
type Plugin interface {
	Cleanup() error
}
