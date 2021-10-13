package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/imagechecker"
)

var (
	skippedInfo = ""

	kymaResourcesDirectory = flag.String("kymaDirectory", "/home/prow/go/src/github.com/kyma-project/kyma/resources/", "Path to Kyma resources")
	skipComments           = flag.Bool("skipComments", true, "Skip commented out lines")
	foundIncompatible      = false
)

func main() {
	flag.Parse()

	if *skipComments {
		skippedInfo = ", excluding commented out lines"
	}

	// for all files in resources
	fmt.Printf("Looking for incompatible images in \"%s\"%s:\n\n", *kymaResourcesDirectory, skippedInfo)

	err := filepath.Walk(*kymaResourcesDirectory, walkFunction)
	if err != nil {
		fmt.Printf("Cannot traverse directory: %s", err)
		os.Exit(2)
	}

	if foundIncompatible {
		fmt.Printf("\nFound incompatible image lines\n")
		os.Exit(3)
	} else {
		fmt.Println("All images seems to bo in the new format")
	}
}

func walkFunction(path string, info fs.FileInfo, err error) error {
	//pass the error further, this shouldn't ever happen
	if err != nil {
		return err
	}

	// skip directory entries, we just want files
	if info.IsDir() {
		return nil
	}

	// we actually only want .yaml files
	if !strings.Contains(info.Name(), ".yaml") {
		return nil
	}

	// TODO move this stuff to its own function, in external file so I can add tests
	incompatible, err := imagechecker.FileHasIncorrectImage(path, *skipComments)
	if err != nil {
		return nil
	}

	if incompatible {
		foundIncompatible = true
	}

	return nil
}
