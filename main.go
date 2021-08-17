package main

import (
	"io/ioutil"

	"github.com/joelanford/declcfg/internal/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetOutput(ioutil.Discard)
	log := logrus.New()
	if err := cmd.New(log).Execute(); err != nil {
		log.Fatal(err)
	}
}
