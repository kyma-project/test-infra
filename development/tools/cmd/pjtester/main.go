package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/development/tools/pkg/pjtester"
	"github.com/sirupsen/logrus"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
)

var (
	log       = logrus.New()
	ghOptions prowflagutil.GitHubOptions
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	if _, present := os.LookupEnv("IMAGE_COMMIT"); present {
		fmt.Printf("IMAGE_COMMIT: %s\n", os.Getenv("IMAGE_COMMIT"))
	}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ghOptions.AddFlags(fs)
	ghOptions.AllowAnonymous = true
	_ = fs.Parse(os.Args[1:])
	if err := ghOptions.Validate(false); err != nil {
		logrus.WithError(err).Fatalf("github options validation failed")
	}
	pjtester.SchedulePJ(ghOptions)
}
