package object

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

type t1 struct {
	Metadata

	T1 string
}

type t2 struct {
	Metadata

	T2 string
}

func newTestRegistry() *Registry {
	reg := Registry{yaml.YamlCodec, make(map[string]Constructor)}
	reg.AddKind("t1", func() BaseObject { return new(t1) })
	reg.AddKind("t2", func() BaseObject { return new(t2) })

	return &reg
}

var (
	testObjs = []BaseObject{
		&t1{
			Metadata{
				Kind: "t1",
				Name: "t1-name",
			},
			"t1p",
		},
		&t2{
			Metadata{
				Kind: "t2",
				Name: "t2-name",
			},
			"t2p",
		},
	}
	testObjsMarshaled = []string{
		`metadata:
  kind: t1
  uid: ""
  generation: 0
  name: t1-name
  namespace: ""
t1: t1p
`,
		`metadata:
  kind: t2
  uid: ""
  generation: 0
  name: t2-name
  namespace: ""
t2: t2p
`,
	}
	testObjsSliceMarshaled = `- metadata:
    kind: t1
    uid: ""
    generation: 0
    name: t1-name
    namespace: ""
  t1: t1p
- metadata:
    kind: t2
    uid: ""
    generation: 0
    name: t2-name
    namespace: ""
  t2: t2p
`
)

func TestRegistry_AddKind(t *testing.T) {
	reg := newTestRegistry()

	assert.Contains(t, reg.kinds, "t1", "Registry should know kind t1")
	assert.Contains(t, reg.kinds, "t2", "Registry should know kind t2")
}

func TestRegistry_MarshalOne(t *testing.T) {
	reg := newTestRegistry()

	data, err := reg.MarshalOne(testObjs[0])
	assert.Nil(t, err, "Object should be marshaled w/o errors")
	assert.Equal(t, testObjsMarshaled[0], string(data), "Correct marshaled data expected")
}

func TestRegistry_MarshalMany(t *testing.T) {
	reg := newTestRegistry()

	data, err := reg.MarshalMany(testObjs)
	assert.Nil(t, err, "Objects should be marshaled w/o errors")
	assert.Equal(t, testObjsSliceMarshaled, string(data), "Correct marshaled data expected")
}

func TestRegistry_UnmarshalOne(t *testing.T) {
	reg := newTestRegistry()

	obj, err := reg.UnmarshalOne([]byte(testObjsMarshaled[0]))
	assert.Nil(t, err, "Object should be unmarshaled w/o errors")
	assert.Exactly(t, testObjs[0], obj, "Unmarshaled object should be deep equal to initial one")
}

func TestRegistry_UnmarshalMany(t *testing.T) {
	reg := newTestRegistry()

	obj, err := reg.UnmarshalMany([]byte(testObjsSliceMarshaled))
	assert.Nil(t, err, "Objects should be unmarshaled w/o errors")
	assert.Exactly(t, testObjs, obj, "Unmarshaled objects should be deep equal to initial ones")
}

const benchmarkSeed = 42

func benchmarkItems(n int) []BaseObject {
	r := rand.New(rand.NewSource(benchmarkSeed))
	items := make([]BaseObject, n)
	for idx := range items {
		items[idx] = &t1{
			Metadata{
				"t1",
				NewUUID(),
				Generation(r.Intn(100)),
				fmt.Sprintf("t1-name-%d-%d", idx, r.Intn(1000000)),
				fmt.Sprintf("t1-namespace-%d-%d", idx, r.Intn(1000000)),
			},
			fmt.Sprintf("t1p-%d-%d", idx, r.Intn(1000000)),
		}
	}

	return items
}

func benchmarkMarshal(b *testing.B, n int) {
	reg := newTestRegistry()

	items := benchmarkItems(n)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if n == 1 {
			if _, err := reg.MarshalOne(items[0]); err != nil {
				b.Fatal(err)
			}
		} else {
			if _, err := reg.MarshalMany(items); err != nil {
				b.Fatal(err)
			}
		}
	}
	b.StopTimer()
}

func BenchmarkRegistry_MarshalOne(b *testing.B)       { benchmarkMarshal(b, 1) }
func BenchmarkRegistry_MarshalMany2(b *testing.B)     { benchmarkMarshal(b, 2) }
func BenchmarkRegistry_MarshalMany10(b *testing.B)    { benchmarkMarshal(b, 10) }
func BenchmarkRegistry_MarshalMany100(b *testing.B)   { benchmarkMarshal(b, 100) }
func BenchmarkRegistry_MarshalMany1000(b *testing.B)  { benchmarkMarshal(b, 1000) }
func BenchmarkRegistry_MarshalMany10000(b *testing.B) { benchmarkMarshal(b, 10000) }

func benchmarkUnmarshal(b *testing.B, n int) {
	reg := newTestRegistry()

	items := benchmarkItems(n)
	var data []byte
	var err error
	if n == 1 {
		data, err = reg.MarshalOne(items[0])
	} else {
		data, err = reg.MarshalMany(items)
	}
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if n == 1 {
			if _, err := reg.UnmarshalOne(data); err != nil {
				b.Fatal(err)
			}
		} else {
			if _, err := reg.UnmarshalMany(data); err != nil {
				b.Fatal(err)
			}
		}
	}
	b.StopTimer()
}

func BenchmarkRegistry_UnmarshalOne(b *testing.B)       { benchmarkUnmarshal(b, 1) }
func BenchmarkRegistry_UnmarshalMany2(b *testing.B)     { benchmarkUnmarshal(b, 2) }
func BenchmarkRegistry_UnmarshalMany10(b *testing.B)    { benchmarkUnmarshal(b, 10) }
func BenchmarkRegistry_UnmarshalMany100(b *testing.B)   { benchmarkUnmarshal(b, 100) }
func BenchmarkRegistry_UnmarshalMany1000(b *testing.B)  { benchmarkUnmarshal(b, 1000) }
func BenchmarkRegistry_UnmarshalMany10000(b *testing.B) { benchmarkUnmarshal(b, 10000) }
