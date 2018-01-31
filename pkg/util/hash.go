package util

import (
	"golang.org/x/crypto/bcrypt"
	"hash/fnv"
)

// HashFnv calculates 32-bit fnv.New32a hash, given a string s
func HashFnv(s string) uint32 {
	hash := fnv.New32a()
	_, err := hash.Write([]byte(s))
	if err != nil {
		panic("Internal error. Can't calculate fnv.New32a() hash from " + s)
	}
	return hash.Sum32()
}

// HashAndSalt returns salted hash from the password (only used to generate user passwords)
func HashAndSalt(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

// ComparePasswords verifies hashed password
func ComparePasswords(hashedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
