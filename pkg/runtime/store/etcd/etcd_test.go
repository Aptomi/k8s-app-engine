package etcd_test

import (
	"testing"

	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"github.com/Aptomi/aptomi/pkg/runtime/store/etcd"
	"github.com/stretchr/testify/assert"
)

func TestEtcdStoreBaseFunctionality(t *testing.T) {
	etcdStore, err := etcd.New(runtime.NewTypes().Append(lang.TypeClaim), store.NewJsonCodec())
	assert.NoError(t, err)
	assert.NotNil(t, etcdStore)

	claim := &lang.Claim{
		TypeKind: lang.TypeClaim.GetTypeKind(),
		Metadata: lang.Metadata{
			Namespace: "some_namespace",
			Name:      "some_name",
		},
		User:    "some_user2",
		Service: "some_service",
		Labels:  map[string]string{},
	}

	err = etcdStore.Save(claim)
	assert.NoError(t, err)

	var loadedClaim *lang.Claim
	err = etcdStore.Find(lang.TypeClaim.Kind /*, WithKey */).Last(loadedClaim)
	assert.NoError(t, err)

	assert.Equal(t, claim, loadedClaim)
}
