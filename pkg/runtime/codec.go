package runtime

type Encoder interface {
	EncodeOne(obj Object) ([]byte, error)
	EncodeMany(objs []Object) ([]byte, error)
}

type Decoder interface {
	DecodeOne(data []byte) (Object, error)
	DecodeOneOrMany(data []byte) ([]Object, error)
}

type Codec interface {
	Encoder
	Decoder
}
