package plugin

type Plugin interface {
	Cleanup() error
}
