package util

import "hash/fnv"

// HashFnv calculates 32-bit fnv.New32a hash, given a string s
func HashFnv(s string) uint32 {
	hash := fnv.New32a()
	_, err := hash.Write([]byte(s))
	if err != nil {
		panic("Internal error. Can't calculate fnv.New32a() hash from " + s)
	}
	return hash.Sum32()
}
