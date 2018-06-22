package dbtest

import (
	"github.com/Aptomi/aptomi/pkg/runtime/db"
)

type TestObj struct {
}

func (*TestObj) GetKind() db.Kind {
	return "test"
}

func (*TestObj) GetKey() db.Key {
	return "key"
}

func init() {
	db.RegisterType(&TestObj{})
}
