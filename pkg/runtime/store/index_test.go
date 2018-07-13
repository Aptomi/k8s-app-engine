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
	assert.NotNil(t, indexes.List)
	assert.Len(t, indexes.List, 3)
	assert.Contains(t, indexes.List, "PolicyGen")
	revision := &engine.Revision{
		TypeKind: engine.TypeRevision.GetTypeKind(),
		Status:   "some_status",
		Metadata: runtime.GenerationMetadata{
			Generation: 1,
		},
		PolicyGen: 42,
	}
	assert.Equal(t, "listgen/system/revision/PolicyGen=42", indexes.NameForStorable("PolicyGen", revision, store.NewJSONCodec()))
	assert.Equal(t, "listgen/system/revision/Status=some_status", indexes.NameForStorable("Status", revision, store.NewJSONCodec()))
	assert.Equal(t, "lastgen/system/revision", indexes.NameForStorable(store.LastGenIndex, revision, store.NewJSONCodec()))

	assert.Equal(t, "listgen/system/revision/PolicyGen=42", indexes.NameForValue("PolicyGen", engine.RevisionKey, 42, store.NewJSONCodec()))
}
