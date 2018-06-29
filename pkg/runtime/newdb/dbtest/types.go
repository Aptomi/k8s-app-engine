package dbtest

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
)

type TestObj struct {
}

func (*TestObj) GetKind() runtime.Kind {
	return "test"
}

func (*TestObj) GetKey() runtime.Key {
	return "key"
}

//func init() {
//	newdb.RegisterType(&TestObj{})
//}
