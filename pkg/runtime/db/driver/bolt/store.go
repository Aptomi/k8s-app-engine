package bolt

import (
	"fmt"
	"sync"

	"github.com/Aptomi/aptomi/pkg/runtime/db"
	"github.com/coreos/bbolt"
)

const (
	boltSeparator = "@"
)

type boltStore struct {
	bolt *bolt.DB
}

type boltTypeInfo struct {
	db.TypeInfo
	objectBucket []byte
	indexBucket  []byte
}

func (store *boltStore) init() error {
	for _, info := range db.GetAllInfos() {
		_, err := store.initForType(info)
		if err != nil {
			return err
		}
	}

	return nil
}

func (store *boltStore) Close() error {
	err := store.bolt.Close()
	if err != nil {
		return fmt.Errorf("error while closing BoltDB: %s", err)
	}

	return nil
}

func (store *boltStore) initForType(dbInfo db.TypeInfo) (*boltTypeInfo, error) {
	// todo optimize by caching already initialized types (in sync.Map? + use sync.Once in type)
	info := &boltTypeInfo{
		TypeInfo:     dbInfo,
		objectBucket: []byte("object" + boltSeparator + dbInfo.Kind()),
		indexBucket:  []byte("index" + boltSeparator + dbInfo.Kind()),
	}

	store.bolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(info.objectBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(info.indexBucket)
		if err != nil {
			return err
		}

		return nil
	})

	return info, nil
}

func (store *boltStore) Update(storable db.Storable, inPlace bool) error {
	info, err := store.initForType(db.GetInfoFor(storable))
	if err != nil {
		return err
	}

	return store.bolt.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(info.objectBucket)

		/*

			flow:
			* get current object from db - if exists, we'll create new version or override


		*/
	})
}

func (store *boltStore) Delete(obj db.Storable, key string) error {
	panic("implement me")
}

func (store *boltStore) Get(result db.Storable, key string) error {
	info, err := store.initForType(db.GetInfoFor(result))
	if err != nil {
		return err
	}

	panic("implement me")
}

func (store *boltStore) List(result []db.Storable, query ...*db.Query) error {
	panic("implement me")
}
