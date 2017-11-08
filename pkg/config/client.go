package config

// Client is the aptomictl config representation
type Client struct {
	Debug bool
	API   API
	Auth  Auth
}

// IsDebug returns true if debug mode enabled
func (c Client) IsDebug() bool {
	return c.Debug
}

// Auth represents client auth configs
type Auth struct {
	Username string `valid:"required"`
}
