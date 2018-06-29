package json

import (
	"encoding/json"

	"github.com/Aptomi/aptomi/pkg/runtime/db"
)

var Codec db.Codec = &codec{}

type codec struct {
}

func (c *codec) Marshal(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c *codec) Unmarshal(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}
