package etcd_test

import (
	"testing"

	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"github.com/Aptomi/aptomi/pkg/runtime/store/etcd"
	"github.com/stretchr/testify/assert"
)

func TestEtcdStoreBaseFunctionality(t *testing.T) {
	etcdStore, err := etcd.New(runtime.NewTypes().Append(engine.TypeRevision), store.NewJsonCodec())
	assert.NoError(t, err)
	assert.NotNil(t, etcdStore)

	revision := &engine.Revision{
		TypeKind: engine.TypeRevision.GetTypeKind(),
		Metadata: runtime.GenerationMetadata{
			Generation: 1,
		},
		PolicyGen: 42,
		Status:    engine.RevisionStatusWaiting,
	}

	err = etcdStore.Save(revision)
	assert.NoError(t, err)

	revision.Status = engine.RevisionStatusCompleted
	err = etcdStore.Save(revision)
	assert.NoError(t, err)

	//var loadedRevisions []*engine.Revision
	err = etcdStore.Find(engine.TypeRevision.Kind, &[]engine.Revision{}, store.WithKey(engine.RevisionKey), store.WithWhereEq("Status", engine.RevisionStatusWaiting, engine.RevisionStatusCompleted))
	assert.NoError(t, err)
	//assert.Len(t, loadedRevisions, 2)

	var loadedRevision *engine.Revision
	err = etcdStore.Find(engine.TypeRevision.Kind, loadedRevision, store.WithKey(engine.RevisionKey), store.WithGen(runtime.LastOrEmptyGen))
	assert.NoError(t, err)
	//assert.Equal(t, revision, loadedRevision)
}
