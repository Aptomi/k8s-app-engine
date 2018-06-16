package bolt

// todo add check for initialized TypeKind here and in the API and to the marshaller

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"github.com/boltdb/bolt"
)

// NewGenericStore creates a new object store based on BoltDB
func NewGenericStore(registry *runtime.Registry) store.Generic {
	codec := yaml.NewCodec(registry)
	return &boltStore{registry: registry, codec: codec}
}

type boltStore struct {
	registry *runtime.Registry
	codec    runtime.Codec
	db       *bolt.DB
}

var objectsBucket = []byte("objects")

func (bs *boltStore) Open(cfg config.DB) error {
	connection := cfg.Connection
	db, err := bolt.Open(connection, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return fmt.Errorf("error while opening BoltDB: %s error: %s", connection, err)
	}
	bs.db = db

	// Initialize all buckets and indexes
	return bs.db.Update(func(tx *bolt.Tx) error {
		_, bucketErr := tx.CreateBucketIfNotExists(objectsBucket)
		return bucketErr
	})
}

func (bs *boltStore) Close() error {
	err := bs.db.Close()
	if err != nil {
		return fmt.Errorf("error while closing BoltDB: %s", err)
	}

	return err
}

const boltSeparator = "@"

func (bs *boltStore) Get(key string) (runtime.Storable, error) {
	var result runtime.Storable
	err := bs.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("bucket not found: %s", objectsBucket)
		}

		data := bucket.Get([]byte(key + boltSeparator + genStr(runtime.LastGen)))

		if data != nil {
			obj, err := bs.codec.DecodeOne(data)
			if err != nil {
				return err
			}
			storable, ok := obj.(runtime.Storable)
			if !ok {
				return fmt.Errorf("storable object is expected to be decoded from bolt, but got: %s", obj.GetKind())
			}
			result = storable
		}

		return nil
	})

	return result, err
}

func (bs *boltStore) GetGen(key string, gen runtime.Generation) (runtime.Versioned, error) {
	var result runtime.Versioned
	err := bs.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("bucket not found: %s", objectsBucket)
		}

		var data []byte
		if gen == runtime.LastGen {
			c := bucket.Cursor()
			prefix := []byte(key + boltSeparator)
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				data = v
			}
		} else {
			data = bucket.Get([]byte(key + boltSeparator + genStr(gen)))
		}

		if data != nil {
			obj, err := bs.codec.DecodeOne(data)
			if err != nil {
				return err
			}
			versioned, ok := obj.(runtime.Versioned)
			if !ok {
				return fmt.Errorf("versioned object is expected to be decoded from bolt, but got: %s", obj.GetKind())
			}
			result = versioned
		}

		return nil
	})

	return result, err
}

func (bs *boltStore) List(prefix string) ([]runtime.Storable, error) {
	result := make([]runtime.Storable, 0)
	err := bs.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("bucket not found: %s", objectsBucket)
		}

		c := bucket.Cursor()
		prefixBytes := []byte(prefix)
		for k, v := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, v = c.Next() {
			baseObj, err := bs.codec.DecodeOne(v)
			if err != nil {
				return err
			}
			obj, ok := baseObj.(runtime.Storable)
			if !ok {
				return fmt.Errorf("storable object is expected to be decoded from bolt, but got: %s", baseObj.GetKind())
			}
			result = append(result, obj)
		}

		return nil
	})

	return result, err
}

func (bs *boltStore) ListGenerations(key string) ([]runtime.Storable, error) {
	return bs.List(key + boltSeparator)
}

func (bs *boltStore) setNextGeneration(obj runtime.Versioned) error {
	// todo replace this code by checking index that returns last generation
	info := bs.registry.Get(obj.GetKind())
	if !info.Versioned {
		return fmt.Errorf("kind %s isn't versioned", obj.GetKind())
	}
	last, err := bs.GetGen(runtime.KeyForStorable(obj), runtime.LastGen)
	if err != nil {
		return err
	}
	var newGen runtime.Generation = 1
	if last != nil {
		newGen = last.GetGeneration().Next()
	}
	obj.SetGeneration(newGen)
	return nil
}

func (bs *boltStore) Save(obj runtime.Storable) (bool, error) {
	return bs.save(obj, false)
}

func (bs *boltStore) Update(obj runtime.Storable) (bool, error) {
	return bs.save(obj, true)
}

func (bs *boltStore) save(obj runtime.Storable, updateCurrent bool) (bool, error) {
	info := bs.registry.Get(obj.GetKind())
	if info == nil {
		return false, fmt.Errorf("unknown kind: %s", obj.GetKind())
	}
	key := runtime.KeyForStorable(obj)
	boltPath := key
	updated := false

	if info.Versioned { // todo extract into "Prepare versioned"
		versionedObj, ok := obj.(runtime.Versioned)
		if !ok {
			return false, fmt.Errorf("versioned object doesn't implement Versioned interface: %s", obj.GetKind())
		}

		// todo we should compare with latest in some cases
		existingObj, err := bs.GetGen(key, versionedObj.GetGeneration())
		if err != nil {
			return false, err
		}

		deletable, ok := obj.(runtime.Deletable)
		if ok && deletable.IsDeleted() {
			if !info.Deletable {
				return false, fmt.Errorf("trying mark object deleted=true that isn't explicitly marked as deletable: %s %s", info.Kind, key)
			}

			if existingObj == nil {
				return false, fmt.Errorf("trying to make non-existing in db object with deleted=true: %s %s", info.Kind, key)
			}
		}

		if existingObj != nil {
			versionedObj.SetGeneration(existingObj.GetGeneration())
			equals, equalsErr := bs.equals(obj, existingObj)
			if equalsErr != nil {
				return false, equalsErr
			}
			if !updateCurrent && !equals {
				errGen := bs.setNextGeneration(versionedObj)
				if errGen != nil {
					return false, fmt.Errorf("error while calling setNextGeneration(%s): %s", obj, errGen)
				}
				updated = true
			}
		} else {
			if versionedObj.GetGeneration() == runtime.LastGen {
				versionedObj.SetGeneration(runtime.FirstGen)
			}
			updated = true
		}

		boltPath += boltSeparator + genStr(versionedObj.GetGeneration())
	} else { // not versioned
		boltPath += boltSeparator + genStr(runtime.LastGen)
	}

	err := bs.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("bucket not found: %s", objectsBucket)
		}

		data, err := bs.codec.EncodeOne(obj)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(boltPath), data)
	})

	return updated, err
}

func (bs *boltStore) Delete(key string) error {
	// todo support deleting version objects, potentially we don't want to remove any object, just mark as deleted

	return bs.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(objectsBucket)
		if bucket == nil {
			return fmt.Errorf("bucket not found: %s", objectsBucket)
		}

		c := bucket.Cursor()
		prefixBytes := []byte(key + boltSeparator)
		for k, v := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, v = c.Next() {
			baseObj, err := bs.codec.DecodeOne(v)
			if err != nil {
				return err
			}
			_, ok := baseObj.(runtime.Versioned)
			if ok {
				return fmt.Errorf("deleting versioned objects isn't implmeneted")
			}

			err = bucket.Delete(k)
			if err != nil {
				return fmt.Errorf("error while deleting object with key: %s", k)
			}
		}

		return nil
	})
}

func (bs *boltStore) equals(o1 runtime.Object, o2 runtime.Object) (bool, error) {
	o1bytes, err := bs.codec.EncodeOne(o1)
	if err != nil {
		return false, err
	}

	o2bytes, err := bs.codec.EncodeOne(o2)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(o1bytes, o2bytes), nil
}

// todo replace with adding bytes to []byte
func genStr(gen runtime.Generation) string {
	return fmt.Sprintf("%20d", gen)
}
