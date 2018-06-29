package dbtest

import (
	"github.com/Aptomi/aptomi/pkg/runtime/db"
)

type TestObj struct {
}

func (*TestObj) GetKind() newdb.Kind {
	return "test"
}

func (*TestObj) GetKey() newdb.Key {
	return "key"
}

func init() {
	newdb.RegisterType(&TestObj{})
}
