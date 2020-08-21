package main

import (
	"fmt"
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
	if _, present := os.LookupEnv("IMAGE_COMMIT"); present {
		fmt.Printf("IMAGE_COMMIT: %s\n", os.Getenv("IMAGE_COMMIT"))
	}
	prtagbuilder.BuildPrTag()
}
