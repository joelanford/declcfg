package main

import (
	"github.com/joelanford/declcfg/internal/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		logrus.Fatal(err)
	}
}
