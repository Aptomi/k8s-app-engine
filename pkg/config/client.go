package config

type Client struct {
	Debug bool
	API   API
	Auth  Auth
	Apply Apply
}

func (c Client) IsDebug() bool {
	return c.Debug
}

type Auth struct {
	Username string `valid:"required"`
}

type Apply struct {
	PolicyPaths []string `valid:"required"`
}
