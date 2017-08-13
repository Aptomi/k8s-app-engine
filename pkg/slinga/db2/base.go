package db2

import (
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"strconv"
	"strings"
)

type Generation uint64

const KeySeparator = "$"

type Key string

func (key Key) parts() []string {
	parts := strings.Split(string(key), "$")
	if len(parts) != 2 {
		panic("Key should consist of two parts separated by " + KeySeparator)
	}
	return parts
}

func (key Key) GetUID() util.UID {
	return util.UID(key.parts()[0])
}

func (key Key) GetGeneration() Generation {
	val, err := strconv.ParseUint(key.parts()[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return Generation(val)
}
