package bolt

import (
	"bytes"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"github.com/boltdb/bolt"
	"io"
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
	last, err := b.GetByName(obj.GetNamespace(), obj.GetKind(), obj.GetName(), object.LastGen)
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

func (b *boltStore) Save(obj object.Base) (bool, error) {
	info := b.catalog.Get(obj.GetKind())
	if info == nil {
		return false, fmt.Errorf("Unknown kind: %s", obj.GetKind())
	}

	updated := false
	if info.Versioned {
		existingObj, err := b.GetByName(obj.GetNamespace(), obj.GetKind(), obj.GetName(), obj.GetGeneration())
		if err != nil {
			return false, err
		}
		if existingObj != nil {
			obj.SetGeneration(existingObj.GetGeneration())
			if !reflect.DeepEqual(obj, existingObj) {
				b.setNextGeneration(obj)
				updated = true
			}
		} else {
			obj.SetGeneration(object.FirstGen)
			updated = true
		}
	} else {
		obj.SetGeneration(object.LastGen)
	}

	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: ")
		}

		data, err := b.codec.MarshalOne(obj)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(strings.Join([]string{obj.GetKey(), obj.GetGeneration().String()}, object.KeySeparator)), data)
	})

	return updated, err
}

func (b *boltStore) GetByName(namespace string, kind string, name string, gen object.Generation) (object.Base, error) {
	// todo support namespaces and kind in different buckets
	var result object.Base
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", objectsBucket)
		}

		var data []byte
		if gen == object.LastGen {
			c := bucket.Cursor()
			prefix := []byte(strings.Join([]string{namespace, kind, name}, object.KeySeparator))
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				data = v
			}
		} else {
			data = bucket.Get([]byte(strings.Join([]string{namespace, kind, name, gen.String()}, object.KeySeparator)))
		}

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

func (b *boltStore) Dump(w io.Writer) error {
	return b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", objectsBucket)
		}

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			_, err := w.Write(v)
			if err != nil {
				return err
			}
			fmt.Fprint(w, "\n====================\n")
		}

		return nil
	})
}
