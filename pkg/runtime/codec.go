package runtime

// Encoder interface represents encoding of the runtime objects into bytes
type Encoder interface {
	EncodeOne(obj Object) ([]byte, error)
	EncodeMany(objs []Object) ([]byte, error)
}

// Decoder interface represents decoding of the runtime objects from bytes
type Decoder interface {
	DecodeOne(data []byte) (Object, error)
	DecodeOneOrMany(data []byte) ([]Object, error)
}

// Codec interface represents combination of Encoder and Decoder interfaces for both sides encoding/decoding of runtime
// objects to/from bytes
type Codec interface {
	Encoder
	Decoder
}
