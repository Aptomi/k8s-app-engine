package etcd

import (
	"context"
	"encoding/binary"

	"github.com/Aptomi/aptomi/pkg/runtime/db"
	"github.com/Aptomi/aptomi/pkg/runtime/db/codec/json"
	etcd "github.com/coreos/etcd/clientv3"
	conc "github.com/coreos/etcd/clientv3/concurrency"
)

type store struct {
	client *etcd.Client
}

func (s *store) Close() error {
	return s.client.Close()
}

func (s *store) Save(new db.Storable, inPlace bool) error {
	info := db.GetInfoFor(new)
	key := new.GetKey()

	// todo seems like we don't need to check response.Success as it's already checked inside the STM
	_, err := conc.NewSTM(s.client, func(stm conc.STM) error {
		// 1. get existing object with desired key and last generation
		// 2. if it exists - deepEqual first => create or update scenario
		// 3. calculate object version => 1 or existing + 1, consider inPlace flag
		// 4. set version for the object
		// 5. save object (Put)
		// 6. iterate over indexes
		// 7. - remove old value from indexes if needed
		// 8. - add new value to indexes if needed

		lastGen := parseIndexSingleGen([]byte(stm.Get(indexPath(info, db.LastGenIndex, key, nil))))
		if lastGen == db.LastOrEmptyGen {
			// creating new object
		}
		current := parseObject([]byte(stm.Get(key + "@" + lastGen.String())))

		// set version for new to the lastGen and encode it, compare results => calculate changed bool
		// consider inPlace, changed

		return nil
	})

	return err
}

func (s *store) Delete(storable db.Storable, key string) error {
	panic("implement me")
}

func (s *store) Get(result db.Storable, key string) error {
	// need to benchmark two gets (last version index + object itself) gets from DB vs sorted get on 10000 version of one object
	// todo replace with two gets, even with second get inside tx it's ~30 times faster then sort
	s.client.Get(context.TODO(), "key@", etcd.WithPrefix(), etcd.WithSort(etcd.SortByKey, etcd.SortDescend), etcd.WithLimit(1))

	panic("implement me")
}

func (s *store) List(result []db.Storable, query ...*db.Query) error {
	panic("implement me")
}

const (
	objectPrefix = "o"
	indexPrefix  = "i"
)

func objectPath(info db.TypeInfo, key db.Key, gen db.Generation) string {
	return objectPrefix + "/" + info.Kind() + "/" + key + "/" + gen.String()
}

func indexPath(info db.TypeInfo, index *db.Index, key db.Key, value interface{}) string {
	//todo
	return indexPrefix + "/" + info.Kind() + "/" + index.Key(key, value)
}

func parseIndexSingleGen(data []byte) db.Generation {
	if len(data) == 0 {
		return db.LastOrEmptyGen
	}

	return db.Generation(binary.BigEndian.Uint64(data))
}

func parseObject(data []byte, result db.Storable) error {
	// todo
	return json.Codec.Unmarshal(data, result)
}
