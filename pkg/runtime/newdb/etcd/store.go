package etcd

import (
	"context"
	"encoding/binary"

	"github.com/Aptomi/aptomi/pkg/runtime/db"
	"github.com/Aptomi/aptomi/pkg/runtime/db/codec/json"
	"github.com/Aptomi/aptomi/pkg/runtime/newdb"
	etcd "github.com/coreos/etcd/clientv3"
	conc "github.com/coreos/etcd/clientv3/concurrency"
)

type store struct {
	client *etcd.Client
}

func (s *store) Close() error {
	return s.client.Close()
}

func (s *store) Save(new newdb.Storable, inPlace bool) error {
	info := newdb.GetInfoFor(new)
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

		lastGen := parseIndexSingleGen([]byte(stm.Get(indexPath(info, newdb.LastGenIndex, key, nil))))
		if lastGen == newdb.LastOrEmptyGen {
			// creating new object
		}
		current := parseObject([]byte(stm.Get(key + "@" + lastGen.String())))

		// set version for new to the lastGen and encode it, compare results => calculate changed bool
		// consider inPlace, changed

		return nil
	})

	return err
}

func (s *store) Delete(storable newdb.Storable, key string) error {
	panic("implement me")
}

func (s *store) Get(result newdb.Storable, key string) error {
	// need to benchmark two gets (last version index + object itself) gets from DB vs sorted get on 10000 version of one object
	// todo replace with two gets, even with second get inside tx it's ~30 times faster then sort
	s.client.Get(context.TODO(), "key@", etcd.WithPrefix(), etcd.WithSort(etcd.SortByKey, etcd.SortDescend), etcd.WithLimit(1))

	panic("implement me")
}

func (s *store) List(result []newdb.Storable, query ...*newdb.Query) error {
	panic("implement me")
}

const (
	objectPrefix = "o"
	indexPrefix  = "i"
)

func objectPath(info newdb.TypeInfo, key newdb.Key, gen newdb.Generation) string {
	return objectPrefix + "/" + info.Kind() + "/" + key + "/" + gen.String()
}

func indexPath(info newdb.TypeInfo, index *newdb.Index, key newdb.Key, value interface{}) string {
	//todo
	return indexPrefix + "/" + info.Kind() + "/" + index.Key(key, value)
}

func parseIndexSingleGen(data []byte) newdb.Generation {
	if len(data) == 0 {
		return newdb.LastOrEmptyGen
	}

	return newdb.Generation(binary.BigEndian.Uint64(data))
}

func parseObject(data []byte, result newdb.Storable) error {
	// todo
	return json.Codec.Unmarshal(data, result)
}
