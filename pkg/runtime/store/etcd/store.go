package etcd

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
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

// todo need to rework keys to not include kind or to start with kind at least???

// Save saves Storable object with specified options into Etcd and updates indexes when appropriate.
// Workflow:
// 1. for non-versioned object key is always static, just put object into etcd and no indexes need to be updated (only
//    generation indexes currently exists)
// 2. for versioned object all manipulations are done inside a single transaction to guarantee atomic operations
//    (like index update, getting last existing generation or comparing with existing object), in addition to that
//    generation set for the object is always ignored if "forceGenOrReplace" option isn't used
// 3. if "replaceOrForceGen" option used, there should be non-zero generation set in the object, last generation will
//    not be checked in that case and old object will be removed from indexes, while new one will be added to them
// 4. default option is saving object with new generation if it differs from the last generation object (or first time
//    created), so, it'll only require adding object to indexes
func (s *etcdStore) Save(newStorable runtime.Storable, opts ...store.SaveOpt) error {
	saveOpts := store.NewSaveOpts(opts)
	info := s.types.Get(newStorable.GetKind())
	indexes := store.IndexesFor(info)
	key := "/" + runtime.KeyForStorable(newStorable)

	if !info.Versioned {
		data := s.marshal(newStorable)
		_, err := s.client.Put(context.TODO(), "/object"+key+"@"+runtime.LastGen.String(), string(data))
		return err
	}

	// todo prefetch all needed keys for STM to maximize performance (in fact it'll get all data in one first request)

	_, err := etcdconc.NewSTM(s.client, func(stm etcdconc.STM) error {
		gen := runtime.FirstGen
		// get index for last gen
		lastGenRaw := stm.Get("/index/" + indexes.KeyForStorable(store.LastGenIndex, newStorable, s.codec))
		if lastGenRaw != "" {
			lastGen, err := strconv.ParseUint(lastGenRaw, 10, 64)
			if err != nil {
				return err
			}
			gen = runtime.Generation(lastGen)

			if !saveOpts.IsReplaceOrForceGen() {
				currObjRaw := stm.Get("/object" + key + "@" + gen.String())

				// todo it's an okay case, we just need to save for the first time
				if currObjRaw == "" {
					// todo better handle
					panic("invalid last gen index (pointing to non existing generation): " + key)
				}
				currObj := info.New().(runtime.Storable)
				s.unmarshal([]byte(currObjRaw), currObj)

				// todo replace deep equals with encoding objects (with same generations) and comparing bytes
				// it should save current gen of second object, set it to first object generation, encode, compare bytes, return back generation
				if !reflect.DeepEqual(currObj, newStorable) {
					gen = gen.Next()
				} else { // todo if not force new version
					//just creating new version
				}
			}
		}

		newStorable.(runtime.Versioned).SetGeneration(gen)
		data := s.marshal(newStorable)
		stm.Put("/object"+key+"@"+gen.String(), string(data))

		// process indexes

		for _, index := range indexes.List {
			indexKey := "/index/" + index.KeyForStorable(newStorable, s.codec)
			if index.Type == store.IndexTypeLastGen {
				// replace format uint with just byte formatting
				stm.Put(indexKey, strconv.FormatUint(uint64(gen), 10))
			} else if index.Type == store.IndexTypeListGen {
				// todo remove from old version in case of replace (b/c old one could become invalid like changed revision status)
				valueList := &store.IndexValueList{}
				valueListRaw := stm.Get(indexKey)
				if valueListRaw != "" {
					s.unmarshal([]byte(valueListRaw), &valueList)
				}
				valueList.Add([]byte(strconv.FormatUint(uint64(gen), 10)))

				data := s.marshal(valueList)

				stm.Put(indexKey, string(data))
			} else {
				panic("only indexes with types store.IndexTypeLastGen and store.IndexTypeListGen are currently supported by Etcd store")
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
