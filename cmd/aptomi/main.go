package main

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/aptomi/root"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	if err := root.Command.Execute(); err != nil {
		panic(fmt.Errorf("error while executing command: %s", err))
	}
}
