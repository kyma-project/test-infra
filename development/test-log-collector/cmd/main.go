package main

import (
	"os"

	logf "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/test-log-collector/cmd/app"
)

func init() {
	logf.SetFormatter(&logf.JSONFormatter{})
	logf.SetOutput(os.Stdout)
}

func main() {
	if err := app.Mainerr(); err != nil {
		logf.Fatal(err)
	}
	logf.Info("success!")
}
