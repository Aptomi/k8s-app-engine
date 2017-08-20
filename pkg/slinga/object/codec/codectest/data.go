package codectest

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
)

type CodecTestObject1 struct {
	Metadata

	Str    string
	Number int
}

type CodecTestObject2 struct {
	Metadata

	Nested CodecTestNestedObject
	Map    map[Kind][]Key
}

type CodecTestNestedObject struct {
	NestedStrs []string
}

var (
	CodecTestObjectsCatalog = &ObjectCatalog{
		map[Kind]*ObjectInfo{
			Kind("t1"): {
				Kind("t1"),
				func() BaseObject { return new(CodecTestObject1) },
			},
			Kind("t2"): {
				Kind("t2"),
				func() BaseObject { return new(CodecTestObject2) },
			},
		},
	}
	CodecTestObjects = []BaseObject{
		&CodecTestObject1{
			Metadata{
				"t1",
				"uid-1",
				1,
				"name-1",
				"namespace-1",
			},
			"str-1",
			1,
		},
		&CodecTestObject1{
			Metadata{
				"t1",
				"uid-2",
				2,
				"name-2",
				"namespace-1",
			},
			"str-2",
			2,
		},
		&CodecTestObject2{
			Metadata{
				"t2",
				"uid-3",
				3,
				"name-3",
				"namespace-2",
			},
			CodecTestNestedObject{
				[]string{"1", "2", "3"},
			},
			map[Kind][]Key{
				Kind("k-1"): {
					KeyFromParts("uid-1", 1),
					KeyFromParts("uid-2", 2),
				},
				Kind("k-2"): {
					KeyFromParts("uid-3", 1),
					KeyFromParts("uid-4", 2),
				},
			},
		},
		&CodecTestObject2{
			Metadata{
				"t2",
				"uid-4",
				4,
				"name-4",
				"namespace-2",
			},
			CodecTestNestedObject{
				[]string{"4", "5", "6"},
			},
			map[Kind][]Key{
				Kind("k-3"): {
					KeyFromParts("uid-5", 1),
					KeyFromParts("uid-6", 2),
				},
				Kind("k-4"): {
					KeyFromParts("uid-7", 1),
					KeyFromParts("uid-8", 2),
				},
			},
		},
	}
)
