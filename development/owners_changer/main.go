package main

import (
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/development/owners_changer/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
