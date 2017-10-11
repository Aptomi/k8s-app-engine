package config

type Client struct {
	Debug bool
	API   API
	Apply Apply
}

func (c Client) IsDebug() bool {
	return c.Debug
}

type Apply struct {
	PolicyPaths []string `valid:"required"`
}
