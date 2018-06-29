package newdb

type Codec interface {
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(data []byte, value interface{}) error
}
