package store

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
)

type ObjectStore interface {
	//TODO pass some config
	Open(connection string) error
	Close() error
	SetCodec(codec codec.MarshalUnmarshaler)
	//GetOne(namespace string, kind Kind, name string, generation Generation) (BaseObject, error)
	//GetOneByKey(key Key) (BaseObject, error)
	GetNewestOne(namespace string, kind string, name string) (BaseObject, error)
	//GetNewestOneByUID(uid UID) (BaseObject, error)
	GetManyByKeys(keys []Key) ([]BaseObject, error)
}

type BaseStore struct {
	Codec codec.MarshalUnmarshaler
}

func (store *BaseStore) SetCodec(codec codec.MarshalUnmarshaler) {
	store.Codec = codec
}
