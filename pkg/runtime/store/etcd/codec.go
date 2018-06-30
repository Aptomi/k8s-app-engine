package etcd

import (
	"fmt"
)

func (s *etcdStore) marshal(value interface{}) []byte {
	data, err := s.codec.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("unable to marshal value %v with error: %s", value, err))
	}

	return data
}

func (s *etcdStore) unmarshal(data []byte, value interface{}) {
	if err := s.codec.Unmarshal(data, value); err != nil {
		panic(fmt.Sprintf("error while unmarshalling data: %s", err))
	}
}
