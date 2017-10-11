package util

import (
	"math/rand"
)

// RandomID generates a random alphanumerical ID, which starts with a letter
func RandomID(rand *rand.Rand, length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		if i == 0 {
			// first letter non-numeric
			b[i] = charset[rand.Intn(len(charset)-10)]
		} else {
			// other letters any
			b[i] = charset[rand.Intn(len(charset))]
		}
	}
	return string(b)
}
