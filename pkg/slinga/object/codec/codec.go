package codec

import "fmt"

// MarshalUnmarshaler represents objects marshaling and unmarshaling such as json, yaml, etc.
type MarshalUnmarshaler interface {
	GetName() string
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(data []byte, value interface{}) error
}

func MarshalUnmarshal(codec MarshalUnmarshaler, from interface{}, to interface{}) {
	data, err := codec.Marshal(from)
	if err != nil {
		panic(fmt.Sprintf("Error while marshaling object: %v", from))
	}
	err = codec.Unmarshal(data, to)
	if err != nil {
		panic(fmt.Sprintf("Error while unmarshaling object %v back from bytes: %v", from, data))
	}
}
