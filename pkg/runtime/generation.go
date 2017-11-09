package runtime

import (
	"fmt"
	"strconv"
)

// Generation represents object's "version" and starts from 0
type Generation uint64

// String returns generation as string to implement Stringer interface
func (gen Generation) String() string {
	return strconv.FormatUint(uint64(gen), 10)
}

// Next returns the next generation of the base object (current + 1)
func (gen Generation) Next() Generation {
	return gen + 1
}

// ParseGeneration returns Generation type representation of specified generation string
func ParseGeneration(gen string) Generation {
	val, err := strconv.ParseUint(gen, 10, 64)
	if err != nil {
		panic(fmt.Errorf("error while parsing generation from %s: %s", gen, err))
	}
	return Generation(val)
}

// GenerationMetadata is the default struct for metadata with only generation in it
type GenerationMetadata struct {
	Generation Generation
}
