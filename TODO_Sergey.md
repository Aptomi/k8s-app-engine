Items to complete:

1. Should we have default "kind" in registry?

1. Do we need created at in metadata?

1. Cache current Revision in Registry (for ultra fast queries to it)
    
1. implement codectest/bench.go with helpers for benchmarking codecs

1. add Yaml benchmarks

1. add Gob codec and benchmarks for it


---
# Initial benchmark impl for codecs

```go
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
```
