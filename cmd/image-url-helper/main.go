package main

import (
	"github.com/kyma-project/test-infra/cmd/image-url-helper/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
# (2025-03-04)