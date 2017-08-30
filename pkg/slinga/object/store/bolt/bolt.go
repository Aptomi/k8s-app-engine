package bolt

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"github.com/boltdb/bolt"
	"time"
)

func NewBoltStore(catalog *object.ObjectCatalog, codec codec.MarshalUnmarshaler) store.ObjectStore {
	return &boltStore{catalog: catalog, codec: codec}
}

type boltStore struct {
	catalog *object.ObjectCatalog
	codec   codec.MarshalUnmarshaler
	db      *bolt.DB
}

func (b *boltStore) Open(connection string) error {
	db, err := bolt.Open(connection, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return fmt.Errorf("Error while opening BoltDB: %s error: %s", connection, err)
	}
	b.db = db

	// Initialize all buckets and indexes
	err = b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(objectsBucket)
		return err
	})

	return nil
}

var objectsBucket = []byte("objects")

func (b *boltStore) Close() error {
	err := b.db.Close()
	if err != nil {
		return fmt.Errorf("Error while closing BoltDB: %s", err)
	}

	return err
}

func (b *boltStore) Save(object object.BaseObject) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: ")
		}

		data, err := b.codec.MarshalOne(object)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(object.GetKey()), data)
		if err != nil {
			return nil
		}

		return nil
	})

	return err
}

func (b *boltStore) GetByKey(key object.Key) (object.BaseObject, error) {
	var result object.BaseObject
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", objectsBucket)
		}

		data := bucket.Get([]byte(key))
		if data != nil {
			obj, err := b.codec.UnmarshalOne(data)
			if err != nil {
				return err
			}
			result = obj
		}

		return nil
	})

	return result, err
}
