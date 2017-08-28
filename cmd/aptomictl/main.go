package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	if err := AptomiCtlCmd.Execute(); err != nil {
		panic(fmt.Sprintf("Error while executing command: %s", err))
	}
}
