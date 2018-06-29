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
	info := s.types.Get(newStorable.GetKind())
	key := "/" + runtime.KeyForStorable(newStorable)

	if !info.Versioned {
		data, err := s.codec.Marshal(newStorable)
		if err != nil {
			return err
		}
		_, err = s.client.Put(context.TODO(), "/object"+key+"@"+runtime.LastGen.String(), string(data))
		return err
	}

	_, err := etcdconc.NewSTM(s.client, func(stm etcdconc.STM) error {
		gen := runtime.FirstGen
		// get index for last gen
		lastGenRaw := stm.Get("/index" + key)
		if lastGenRaw != "" {
			gen = runtime.Generation(binary.BigEndian.Uint64([]byte(lastGenRaw)))

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
				return nil
			}
		}

		// todo make wrapper that will panic as it's ok to panic if can't marshal/unmarshal data
		// todo just have defer recover at the beginning of each function...
		newStorable.SetGeneration(gen)
		data, cErr := s.codec.Marshal(newStorable)
		if cErr != nil {
			return cErr
		}
		stm.Put("/object"+key+"@"+gen.String(), string(data))

		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(gen))
		stm.Put("/index"+key, string(buf))

		return nil
	})

	return err
}

func (s *etcdStore) Find(kind runtime.Kind, opts ...store.FindOpt) store.Finder {
	// todo next 3 lines could be extracted to the common code
	findOpts := &store.FindOpts{}
	for _, opt := range opts {
		opt(findOpts)
	}

	return &finder{s, kind, findOpts}
}

func (s *etcdStore) Delete(kind runtime.Kind, opts ...store.DeleteOpt) store.Deleter {
	panic("implement me")
}

type finder struct {
	*etcdStore
	kind runtime.Kind
	*store.FindOpts
}

func (f *finder) First(runtime.Storable) error {
	panic("implement me")
}

func (f *finder) Last(runtime.Storable) error {
	panic("implement me")
}

func (f *finder) List(interface{}) error {
	panic("implement me")
}
