package main

import (
	"log"

	"github.com/kyma-project/test-infra/development/image-detector/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("failed to run command: %s", err)
	}
}
