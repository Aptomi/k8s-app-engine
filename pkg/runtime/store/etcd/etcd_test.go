package etcd_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"github.com/Aptomi/aptomi/pkg/runtime/store/etcd"
	"github.com/stretchr/testify/assert"
)

func TestEtcdStoreBaseFunctionality(t *testing.T) {
	// todo create helper in future that'll allow to run any number of tests in parallel - some random constant generated on start + test name as prefix
	endpoints := os.Getenv("APTOMI_TEST_DB_ENDPOINTS")
	if endpoints == "" {
		endpoints = "127.0.0.1:2379"
	}
	cfg := etcd.Config{
		Prefix:    t.Name(),
		Endpoints: strings.Split(endpoints, ","),
	}
	// todo test with all codecs
	etcdStore, err := etcd.New(cfg, runtime.NewTypes().Append(engine.TypeRevision), store.NewGobCodec())
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

	var changed bool
	changed, err = etcdStore.Save(revision)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.EqualValues(t, revision.GetGeneration(), 1)

	revision.Status = engine.RevisionStatusInProgress
	changed, err = etcdStore.Save(revision)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.EqualValues(t, revision.GetGeneration(), 2)

	changed, err = etcdStore.Save(revision)
	assert.NoError(t, err)
	assert.False(t, changed)
	assert.EqualValues(t, revision.GetGeneration(), 2)

	var loadedRevisions []*engine.Revision
	err = etcdStore.Find(engine.TypeRevision.Kind, &loadedRevisions, store.WithKey(engine.RevisionKey), store.WithWhereEq("Status", engine.RevisionStatusWaiting, engine.RevisionStatusInProgress))
	assert.NoError(t, err)
	assert.Len(t, loadedRevisions, 2)
	assert.NotNil(t, loadedRevisions[0])
	assert.NotNil(t, loadedRevisions[1])
	assert.Equal(t, engine.RevisionStatusWaiting, loadedRevisions[0].Status)
	assert.EqualValues(t, 1, loadedRevisions[0].GetGeneration())
	assert.Equal(t, engine.RevisionStatusInProgress, loadedRevisions[1].Status)
	assert.EqualValues(t, 2, loadedRevisions[1].GetGeneration())

	var loadedRevisionByLastGen *engine.Revision
	err = etcdStore.Find(engine.TypeRevision.Kind, &loadedRevisionByLastGen, store.WithKey(engine.RevisionKey), store.WithGen(runtime.LastOrEmptyGen))
	assert.NoError(t, err)
	assert.Equal(t, revision, loadedRevisionByLastGen)

	var loadedRevisionBySpecificGen *engine.Revision
	err = etcdStore.Find(engine.TypeRevision.Kind, &loadedRevisionBySpecificGen, store.WithKey(engine.RevisionKey), store.WithGen(2))
	assert.NoError(t, err)
	assert.Equal(t, revision, loadedRevisionBySpecificGen)

	err = etcdStore.Find(engine.TypeRevision.Kind, &loadedRevisionBySpecificGen, store.WithKey(engine.RevisionKey), store.WithGen(42))
	assert.NoError(t, err)
	assert.Nil(t, loadedRevisionBySpecificGen)
}
