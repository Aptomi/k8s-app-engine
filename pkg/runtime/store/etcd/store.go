package etcd

import (
	"context"
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	etcd "github.com/coreos/etcd/clientv3"
	etcdconc "github.com/coreos/etcd/clientv3/concurrency"
)

type etcdStore struct {
	client *etcd.Client
	types  *runtime.Types
	codec  store.Codec
}

func New( /* todo config */ types *runtime.Types, codec store.Codec) (store.Interface, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("error while connecting to etcd: %s", err)
	}

	// todo run compactor?

	return &etcdStore{
		client: client,
		types:  types,
		codec:  codec,
	}, nil
}

func (s *etcdStore) Close() error {
	return s.client.Close()
}

// todo need to rework keys to not include kind or to start with kind at least

func (s *etcdStore) Save(newStorable runtime.Storable, opts ...store.SaveOpt) error {
	saveOpts := store.NewSaveOpts(opts)
	info := s.types.Get(newStorable.GetKind())
	indexes := store.Indexes(info)
	key := "/" + runtime.KeyForStorable(newStorable)

	if !info.Versioned {
		data, err := s.codec.Marshal(newStorable)
		if err != nil {
			return err
		}
		_, err = s.client.Put(context.TODO(), "/object"+key+"@"+runtime.LastGen.String(), string(data))
		return err
	}

	/*
		What is index? GenIndex is object key + optional field

		* last gen used: for each key there should be last gen stored: key -> gen
		* list revision gens for policy: key+policy -> gen1,gen2,gen3 (in case of revisions key is static)
		* first / last revision gen for policy: just use list revision gen for policy index

	*/

	_, err := etcdconc.NewSTM(s.client, func(stm etcdconc.STM) error {
		gen := runtime.FirstGen
		// get index for last gen
		lastGenRaw := stm.Get("/index" + key)
		if lastGenRaw != "" {
			gen = runtime.Generation(binary.BigEndian.Uint64([]byte(lastGenRaw)))

			if !saveOpts.IsReplace() {
				currObjRaw := stm.Get("/object" + key + "@" + gen.String())
				if currObjRaw == "" {
					// todo better handle
					panic("invalid last gen index (pointing to non existing generation): " + key)
				}
				currObj := info.New().(runtime.Storable)
				err := s.codec.Unmarshal([]byte(currObjRaw), currObj)
				if err != nil {
					return err
				}

				if !reflect.DeepEqual(currObj, newStorable) {
					gen = gen.Next()
				} else { // todo if not force new version
					//just creating new version
				}
			}
		}

		// todo make wrapper that will panic as it's ok to panic if can't marshal/unmarshal data
		// todo just have defer recover at the beginning of each function...
		newStorable.(runtime.Versioned).SetGeneration(gen)
		data, cErr := s.codec.Marshal(newStorable)
		if cErr != nil {
			return cErr
		}
		stm.Put("/object"+key+"@"+gen.String(), string(data))

		// process indexes

		for _, index := range indexes {
			if index.Type == store.IndexTypeLast {
				// we will need to delete old value from old index if we want to support more Last indexes
				//newIndexKey := index.KeyForStorable(newStorable, s.codec)
			}
		}

		return nil
	})

	return err
}

/*
Current Find use cases:

* Find(kind).List
* Find(kind, key).One (non-versioned, should set version to 0 and delegate to next)
* Find(kind, key, gen).One (versioned)

* Find(engine.TypeRevision.Kind, store.WithKey(engine.RevisionKey), store.WithWhereEq("PolicyGen", policyGen), store.WithGetLast()).One(revision)
* Find(engine.TypeRevision.Kind, store.WithKey(engine.RevisionKey), store.WithWhereEq("PolicyGen", policyGen)).List(&revisions)
* Find(engine.TypeRevision.Kind, store.WithKey(engine.RevisionKey), store.WithWhereEq("Status", engine.RevisionStatusWaiting, engine.RevisionStatusInProgress), store.WithGetFirst()).One(revision)
 *Find(engine.TypeDesiredState.Kind, store.WithKey(runtime.KeyFromParts(runtime.SystemNS, engine.TypeDesiredState.Kind, engine.GetDesiredStateName(revision.GetGeneration())))).One(desiredState)

*/

func (s *etcdStore) Find(kind runtime.Kind, opts ...store.FindOpt) store.Finder {
	findOpts := store.NewFindOpts(opts)

	if findOpts.GetKey() == "" {
		// todo handle as separated case to return all objects for specified kind
	}

	info := s.types.Get(kind)

	return &finder{s, findOpts, info}
}

type finder struct {
	*etcdStore
	*store.FindOpts
	info *runtime.TypeInfo
}

func (f *finder) One(runtime.Storable) error {
	panic("implement me")
}

func (f *finder) List(interface{}) error {
	panic("implement me")
}

func (s *etcdStore) Delete(kind runtime.Kind, key runtime.Key) error {
	panic("implement me")
}
