package fake

import "errors"

type errCodec struct {
}

// ErrCodecName is the name of Err MarshalUnmarshaler implementation (always returns err)
const ErrCodecName = "err"

// ErrCodec is the instance of Err MarshalUnmarshaler (always returns err)
var ErrCodec = errCodec{}

func (c errCodec) GetName() string {
	return ErrCodecName
}

func (c errCodec) Marshal(value interface{}) ([]byte, error) {
	return make([]byte, 0), errors.New("Marshal error")
}

func (c errCodec) Unmarshal(data []byte, value interface{}) error {
	return errors.New("Unmarshal error")
}
