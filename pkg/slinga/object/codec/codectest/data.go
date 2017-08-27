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
		Infos: map[Kind]*ObjectInfo{
			Kind("t1"): {
				Kind:        Kind("t1"),
				Constructor: func() BaseObject { return new(CodecTestObject1) },
			},
			Kind("t2"): {
				Kind:        Kind("t2"),
				Constructor: func() BaseObject { return new(CodecTestObject2) },
			},
		},
	}
	CodecTestObjects = []BaseObject{
		&CodecTestObject1{
			Metadata: Metadata{
				Kind:       "t1",
				UID:        "uid-1",
				Generation: 1,
				Name:       "name-1",
				Namespace:  "namespace-1",
			},
			Str:    "str-1",
			Number: 1,
		},
		&CodecTestObject1{
			Metadata: Metadata{
				Kind:       "t1",
				UID:        "uid-2",
				Generation: 2,
				Name:       "name-2",
				Namespace:  "namespace-1",
			},
			Str:    "str-2",
			Number: 2,
		},
		&CodecTestObject2{
			Metadata: Metadata{
				Kind:       "t2",
				UID:        "uid-3",
				Generation: 3,
				Name:       "name-3",
				Namespace:  "namespace-2",
			},
			Nested: CodecTestNestedObject{
				NestedStrs: []string{"1", "2", "3"},
			},
			Map: map[Kind][]Key{
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
			Metadata: Metadata{
				Kind:       "t2",
				UID:        "uid-4",
				Generation: 4,
				Name:       "name-4",
				Namespace:  "namespace-2",
			},
			Nested: CodecTestNestedObject{
				NestedStrs: []string{"4", "5", "6"},
			},
			Map: map[Kind][]Key{
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
