package store_test

import (
	"testing"

	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	"github.com/stretchr/testify/assert"
)

func TestIndexes(t *testing.T) {
	indexes := store.IndexesFor(engine.TypeRevision)
	assert.NotNil(t, indexes)
	assert.Len(t, indexes, 3)
	assert.Contains(t, indexes, "PolicyGen")
	revision := &engine.Revision{
		TypeKind: engine.TypeRevision.GetTypeKind(),
		Status:   "some_status",
		Metadata: runtime.GenerationMetadata{
			Generation: 1,
		},
		PolicyGen: 42,
	}
	assert.Equal(t, "system/revision@PolicyGen@42", indexes.KeyForStorable("PolicyGen", revision, store.NewJsonCodec()))
	assert.Equal(t, "system/revision@Status@some_status", indexes.KeyForStorable("Status", revision, store.NewJsonCodec()))
	assert.Equal(t, "system/revision", indexes.KeyForStorable(store.LastGenIndex, revision, store.NewJsonCodec()))

	assert.Equal(t, "system/revision@PolicyGen@42", indexes.KeyForValue("PolicyGen", engine.RevisionKey, 42, store.NewJsonCodec()))
}
