package main

import (
	"fmt"
	"github.com/Aptomi/aptomi/cmd/aptomictl/root"
	"github.com/Sirupsen/logrus"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	defer func() {
		if err := recover(); err != nil {
			logrus.Fatalf("%s", err)
		}
	}()

	if err := root.Command.Execute(); err != nil {
		panic(fmt.Errorf("error while executing command: %s", err))
	}
}
