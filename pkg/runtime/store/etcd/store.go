package etcd

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	etcd "github.com/coreos/etcd/clientv3"
	etcdconc "github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/clientv3/namespace"
)

type etcdStore struct {
	client *etcd.Client
	types  *runtime.Types
	codec  store.Codec
}

func New(cfg Config, types *runtime.Types, codec store.Codec) (store.Interface, error) {
	if len(cfg.Endpoints) == 0 {
		cfg.Endpoints = []string{"localhost:2379"}
	}

	client, err := etcd.New(etcd.Config{
		Endpoints:            cfg.Endpoints,
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("error while connecting to etcd: %s", err)
	}

	cfg.Prefix = strings.Trim(cfg.Prefix, "/")
	if cfg.Prefix != "" {
		cfg.Prefix = "/" + cfg.Prefix
		client.KV = namespace.NewKV(client.KV, cfg.Prefix)
		client.Lease = namespace.NewLease(client.Lease, cfg.Prefix)
		client.Watcher = namespace.NewWatcher(client.Watcher, cfg.Prefix)
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
func (s *etcdStore) Save(newStorable runtime.Storable, opts ...store.SaveOpt) (bool, error) {
	if newStorable == nil {
		return false, fmt.Errorf("can't save nil")
	}

	saveOpts := store.NewSaveOpts(opts)
	info := s.types.Get(newStorable.GetKind())
	indexes := store.IndexesFor(info)
	key := "/" + runtime.KeyForStorable(newStorable)

	if !info.Versioned {
		data := s.marshal(newStorable)
		_, err := s.client.KV.Put(context.TODO(), "/object"+key+"@"+runtime.LastOrEmptyGen.String(), string(data))
		// todo should it be true or false always?
		return false, err
	}

	var newVersion bool
	newObj := newStorable.(runtime.Versioned)
	// todo prefetch all needed keys for STM to maximize performance (in fact it'll get all data in one first request)
	// todo consider unmarshal to the info.New() to support gob w/o need to register types?
	_, err := etcdconc.NewSTM(s.client, func(stm etcdconc.STM) error {
		// need to remove this obj from indexes
		var prevObj runtime.Storable

		if saveOpts.IsReplaceOrForceGen() {
			newGen := newObj.GetGeneration()
			if newGen == runtime.LastOrEmptyGen {
				return fmt.Errorf("error while saving object %s with replaceOrForceGen option but with empty generation", key)
			}
			// need to check if there is an object already exists with gen from the object, if yes - remove it from indexes
			oldObjRaw := stm.Get("/object" + key + "@" + newGen.String())
			if oldObjRaw != "" {
				// todo avoid
				prevObj = info.New().(runtime.Storable)
				/*
					add field require not nil val for unmarshal field into codec
					if nil passed => create instance of desired object (w/o casting to storable) and pass to unmarshal
					if not nil => error if incorrect type
				*/
				s.unmarshal([]byte(oldObjRaw), prevObj)
			}

			// todo compare - if not changed - nothing to do
		} else {
			// need to get last gen using index, if exists - compare with, if different - increment revision and delete old from indexes
			lastGenRaw := stm.Get("/index/" + indexes.KeyForStorable(store.LastGenIndex, newStorable, s.codec))
			if lastGenRaw == "" {
				newObj.SetGeneration(runtime.FirstGen)
				newVersion = true
			} else {
				lastGen := s.unmarshalGen(lastGenRaw)
				oldObjRaw := stm.Get("/object" + key + "@" + lastGen.String())
				if oldObjRaw == "" {
					return fmt.Errorf("last gen index for %s seems to be corrupted: generation doesn't exist", key)
				}
				// todo avoid
				prevObj = info.New().(runtime.Storable)
				s.unmarshal([]byte(oldObjRaw), prevObj)
				newObj.SetGeneration(lastGen)
				if reflect.DeepEqual(prevObj, newObj) {
					return nil
				} else {
					newObj.SetGeneration(lastGen.Next())
					newVersion = true
				}
			}
		}

		data := s.marshal(newObj)
		newGen := newObj.GetGeneration()
		stm.Put("/object"+key+"@"+newGen.String(), string(data))

		if prevObj != nil && prevObj.(runtime.Versioned).GetGeneration() == newGen {
			for _, index := range indexes.List {
				indexKey := "/index/" + index.KeyForStorable(prevObj, s.codec)
				if index.Type == store.IndexTypeListGen {
					s.updateIndex(stm, indexKey, prevObj.(runtime.Versioned).GetGeneration(), true)
				}
			}
		}

		for _, index := range indexes.List {
			indexKey := "/index/" + index.KeyForStorable(newStorable, s.codec)
			if index.Type == store.IndexTypeLastGen {
				stm.Put(indexKey, s.marshalGen(newGen))
			} else if index.Type == store.IndexTypeListGen {
				s.updateIndex(stm, indexKey, newGen, false)
			} else {
				panic("only indexes with types store.IndexTypeLastGen and store.IndexTypeListGen are currently supported by Etcd store")
			}
		}

		return nil
	})

	return newVersion, err
}

func (s *etcdStore) updateIndex(stm etcdconc.STM, indexKey string, newGen runtime.Generation, delete bool) {
	valueList := &store.IndexValueList{}
	valueListRaw := stm.Get(indexKey)
	if valueListRaw != "" {
		s.unmarshal([]byte(valueListRaw), valueList)
	}
	// todo avoid marshaling gens for indexes by using special index value list type for gens
	gen := []byte(s.marshalGen(newGen))
	if delete {
		valueList.Remove(gen)
	} else {
		valueList.Add(gen)
	}
	data := s.marshal(valueList)
	stm.Put(indexKey, string(data))
}

/*
Current Find use cases:

* Find(kind, keyPrefix)
* Find(kind, key, gen)  (gen=0 for non-versioned)
* Find(kind, key, WithWhereEq)
* Find(kind, key, WithWhereEq, WithGetFirst)
* Find(kind, key, WithWhereEq, WithGetLast)

\\ summary: keyPrefix OR key+gen OR key + whereEq+list/first/last

Workflow:
* validate parameters and result
* identify requested list or one(first or last)
* build list of keys that are result (could be just build key from parameters or use index)
* based on requested list/first/last get corresponding element from the key list and query value for it

*/
func (s *etcdStore) Find(kind runtime.Kind, result interface{}, opts ...store.FindOpt) error {
	findOpts := store.NewFindOpts(opts)
	info := s.types.Get(kind)

	resultTypeElem := reflect.TypeOf(info.New())
	resultTypeSingle := reflect.PtrTo(reflect.TypeOf(info.New()))
	resultTypeList := reflect.PtrTo(reflect.SliceOf(resultTypeElem))

	resultList := false

	resultType := reflect.TypeOf(result)
	if resultType == resultTypeSingle {
		// ok!
	} else if resultType == resultTypeList {
		// ok!
		resultList = true
	} else {
		// todo return back verification
		fmt.Printf("result should be %s or %s, but found: %s\n", resultTypeSingle, resultTypeList, resultType)
		//return fmt.Errorf("result should be %s or %s, but found: %s", resultTypeSingle, resultTypeList, resultType)
	}

	v := reflect.ValueOf(result).Elem()
	if findOpts.GetKeyPrefix() != "" {
		return s.findByKeyPrefix(findOpts, info, func(elem interface{}) {
			// todo validate type of the elem
			// todo if !resultList
			v.Set(reflect.Append(v, reflect.ValueOf(elem)))
		})
	} else if findOpts.GetKey() != "" && findOpts.GetFieldEqName() == "" {
		return s.findByKey(findOpts, info, func(elem interface{}) {
			// todo validate type of the elem
			if elem == nil {
				v.Set(reflect.Zero(v.Type()))
			} else {
				v.Set(reflect.ValueOf(elem))
			}
		})
	} else {
		return s.findByFieldEq(findOpts, info, func(elem interface{}) {
			// todo validate type of the elem
			if !resultList {
				if elem == nil {
					v.Set(reflect.Zero(v.Type()))
				} else {
					v.Set(reflect.ValueOf(elem))
				}
			} else {
				v.Set(reflect.Append(v, reflect.ValueOf(elem)))
			}
		})
	}
}

func (s *etcdStore) findByKeyPrefix(findOpts *store.FindOpts, info *runtime.TypeInfo, addToResult func(interface{})) error {
	if info.Versioned {
		return fmt.Errorf("searching with key prefix is only supported for non versioned objects")
	}

	resp, err := s.client.KV.Get(context.TODO(), "/object"+"/"+findOpts.GetKeyPrefix(), etcd.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		// todo avoid
		elem := info.New()
		s.unmarshal(kv.Value, elem)
		addToResult(elem)
	}

	return nil
}

func (s *etcdStore) findByKey(findOpts *store.FindOpts, info *runtime.TypeInfo, addToResult func(interface{})) error {

	if !info.Versioned && findOpts.GetGen() != runtime.LastOrEmptyGen {
		return fmt.Errorf("requested specific version for non versioned object")
	}

	var data []byte

	if !info.Versioned || findOpts.GetGen() != runtime.LastOrEmptyGen {
		resp, respErr := s.client.KV.Get(context.TODO(), "/object"+"/"+findOpts.GetKey()+"@"+findOpts.GetGen().String())
		if respErr != nil {
			return respErr
		} else if resp.Count > 0 {
			data = resp.Kvs[0].Value
		}
	} else {
		indexes := store.IndexesFor(info)
		// todo wrap into STM to ensure we're getting really last unchanged element / consider is it important? we can't delete generation, so, probably no need for STM here
		resp, respErr := s.client.KV.Get(context.TODO(), "/index/"+indexes.KeyForValue(store.LastGenIndex, findOpts.GetKey(), nil, s.codec))
		if respErr != nil {
			return respErr
		} else if resp.Count > 0 {
			lastGen := s.unmarshalGen(string(resp.Kvs[0].Value))
			resp, respErr = s.client.KV.Get(context.TODO(), "/object"+"/"+findOpts.GetKey()+"@"+lastGen.String())
			if respErr != nil {
				return respErr
			} else if resp.Count > 0 {
				data = resp.Kvs[0].Value
			}
		}
	}

	if data == nil {
		addToResult(nil)
	} else {
		// todo avoid
		result := info.New()
		s.unmarshal(data, result)

		addToResult(result)
	}

	return nil
}

func (s *etcdStore) findByFieldEq(findOpts *store.FindOpts, info *runtime.TypeInfo, addToResult func(interface{})) error {
	indexes := store.IndexesFor(info)
	resultGens := make([]runtime.Generation, 0)

	_, err := etcdconc.NewSTM(s.client, func(stm etcdconc.STM) error {
		for _, fieldValue := range findOpts.GetFieldEqValues() {
			indexKey := "/index/" + indexes.KeyForValue(findOpts.GetFieldEqName(), findOpts.GetKey(), fieldValue, s.codec)
			indexValue := stm.Get(indexKey)
			if indexValue != "" {
				valueList := &store.IndexValueList{}
				s.unmarshal([]byte(indexValue), valueList)
				for _, val := range *valueList {
					resultGens = append(resultGens, s.unmarshalGen(string(val)))
				}
			}
		}

		sort.Slice(resultGens, func(i, j int) bool {
			return resultGens[i] < resultGens[j]
		})

		if len(resultGens) > 0 {
			if findOpts.IsGetFirst() {
				resultGens = []runtime.Generation{resultGens[0]}
			} else if findOpts.IsGetLast() {
				resultGens = []runtime.Generation{resultGens[len(resultGens)-1]}
			}
			for _, gen := range resultGens {
				data := stm.Get("/object" + "/" + findOpts.GetKey() + "@" + gen.String())
				if data == "" {
					return fmt.Errorf("index is invalid :(")
				}
				result := info.New()
				s.unmarshal([]byte(data), result)
				addToResult(result)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *etcdStore) Delete(kind runtime.Kind, key runtime.Key) error {
	info := s.types.Get(kind)

	if info.Versioned {
		return fmt.Errorf("versioned object couldn't be deleted using store.Delete, use deleted flag + store.Save instead")
	}

	_, err := s.client.KV.Delete(context.TODO(), "/object"+"/"+key+"@"+runtime.LastOrEmptyGen.String())

	return err
}
