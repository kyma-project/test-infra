package main

import (
	"os"

	"github.com/kyma-project/test-infra/development/tools/pkg/prtagbuilder"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	prtagbuilder.BuildPrTag()
}
