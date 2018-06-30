package etcd

import (
	"encoding/binary"
	"fmt"

	"github.com/Aptomi/aptomi/pkg/runtime"
)

func (s *etcdStore) marshal(value interface{}) []byte {
	data, err := s.codec.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("error while marshaling value %v with error: %s", value, err))
	}

	return data
}

func (s *etcdStore) unmarshal(data []byte, value interface{}) {
	if err := s.codec.Unmarshal(data, value); err != nil {
		panic(fmt.Sprintf("error while unmarshaling data: %s", err))
	}
}

func (s *etcdStore) marshalGen(generation runtime.Generation) string {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(generation))

	return string(data)
}

func (s *etcdStore) unmarshalGen(data string) runtime.Generation {
	return runtime.Generation(binary.BigEndian.Uint64([]byte(data)))
}
