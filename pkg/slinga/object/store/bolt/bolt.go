package bolt

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"github.com/boltdb/bolt"
	"reflect"
	"strings"
	"time"
)

func NewBoltStore(catalog *object.Catalog, codec codec.MarshalUnmarshaler) store.ObjectStore {
	return &boltStore{catalog: catalog, codec: codec}
}

type boltStore struct {
	catalog *object.Catalog
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

func (b *boltStore) setNextGeneration(obj object.Base) error {
	// todo replace this code by checking index that returns last generation
	info := b.catalog.Get(obj.GetKind())
	if !info.Versioned {
		return fmt.Errorf("Kind %s isn't versioned", obj.GetKind())
	}
	last, err := b.GetByKey(obj.GetNamespace(), obj.GetKind(), obj.GetKey(), object.LastGen)
	if err != nil {
		return err
	}
	var newGen object.Generation = 1
	if last != nil {
		newGen = last.GetGeneration().Next()
	}
	obj.SetGeneration(newGen)
	return nil
}

func (b *boltStore) Save(obj object.Base) (updated bool, err error) {
	info := b.catalog.Get(obj.GetKind())
	if info.Versioned {
		existingObj, err := b.GetByKey(obj.GetNamespace(), obj.GetKind(), obj.GetKey(), obj.GetGeneration())
		if err != nil {
			return false, err
		}
		if !reflect.DeepEqual(obj, existingObj) {
			b.setNextGeneration(obj)
			updated = true
		}
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: ")
		}

		data, err := b.codec.MarshalOne(obj)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(strings.Join([]string{obj.GetNamespace(), obj.GetKind(), obj.GetKey()}, "_")), data)
		if err != nil {
			return err
		}

		return nil
	})

	return updated, err
}

func (b *boltStore) GetByKey(namespace string, kind string, key string, gen object.Generation) (object.Base, error) {
	// todo support namespaces and kind in different buckets
	// todo support generations
	var result object.Base
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", objectsBucket)
		}

		data := bucket.Get([]byte(strings.Join([]string{namespace, kind, key}, "_")))
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
