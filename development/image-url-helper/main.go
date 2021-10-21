package main

import (
	"os"

	"github.com/kyma-project/test-infra/development/image-url-helper/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
