package object

import "math/rand"

const (
	randAddonLength = 6
	letterBytes     = "abcdefghijklmnopqrstuvwxyz0123456789" // all letters to be used in RandAddon
	letterIdxBits   = 6                                      // 6 bits to represent a letter index
	letterIdxMask   = 1<<letterIdxBits - 1                   // letter idx mask (letterIdxBits of 1)
	letterIdxMax    = 63 / letterIdxBits                     // number of letter indices in 63 bits
)

func init() {
	// check that at least one rand addon could be generated out of the single random 63 bits
	if len(letterBytes) > 63 {
		panic("Number of source letters should be less or equal then size of letter idx (letterIdxBits)")
	}
}

func NewRandAddon() string {
	bytes := make([]byte, randAddonLength)
	for letterIdx, random, remain := 0, int64(0), 0; letterIdx < randAddonLength; {
		if remain == 0 {
			random, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(random & letterIdxMask); idx < len(letterBytes) {
			bytes[letterIdx] = letterBytes[idx]
			letterIdx++
		}
		random >>= letterIdxBits
		remain--
	}

	return string(bytes)
}
