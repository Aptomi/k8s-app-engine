package gob

import (
	"bytes"
	"encoding/gob"

	"github.com/Aptomi/aptomi/pkg/runtime/db"
)

var Codec newdb.Codec = &codec{}

type codec struct {
}

func (c *codec) Marshal(value interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(value)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c *codec) Unmarshal(data []byte, value interface{}) error {
	var buffer bytes.Buffer
	decoder := gob.NewDecoder(&buffer)

	_, err := buffer.Write(data)
	if err != nil {
		return err
	}

	return decoder.Decode(value)
}
