package main

import (
	"github.com/Aptomi/aptomi/cmd/aptomictl/root"
	"github.com/Sirupsen/logrus"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	defer func() {
		if err := recover(); err != nil {
			logrus.Fatalf("%s", err)
			os.Exit(1)
		}
	}()

	if err := root.Command.Execute(); err != nil {
		logrus.Fatalf("%s", err)
		os.Exit(1)
	}
}
