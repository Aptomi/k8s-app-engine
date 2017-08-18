package codec

// MarshalUnmarshaler represents objects marshaling and unmarshaling such as json, yaml, etc.
type MarshalUnmarshaler interface {
	GetName() string
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(data []byte, value interface{}) error
}
