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
	kymaResourcesDirectory = flag.String("kymaDirectory", "/home/prow/go/src/github.com/kyma-project/kyma/resources/", "Path to Kyma resources")
	skipComments           = flag.Bool("skipComments", true, "Skip commented out lines")
)

func main() {
	flag.Parse()
	foundIncompatible := false

	skipComentsInfo := ""
	if *skipComments {
		skipComentsInfo = ", excluding commented out lines"
	}

	// for all files in resources
	fmt.Printf("Looking for incompatible images in \"%s\"%s:\n\n", *kymaResourcesDirectory, skipComentsInfo)

	err := filepath.Walk(*kymaResourcesDirectory, getWalkFunc(&foundIncompatible))
	if err != nil {
		fmt.Printf("Cannot traverse directory: %s", err)
		os.Exit(2)
	}

	if foundIncompatible {
		fmt.Printf("\nFound incompatible image lines\n")
		os.Exit(3)
	} else {
		fmt.Println("All images seems to be in the new format")
	}
}

func getWalkFunc(foundIncompatible *bool) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		//pass the error further, this shouldn't ever happen
		if err != nil {
			return err
		}

		// skip directory entries, we just want files
		if info.IsDir() {
			return nil
		}

		// we only want to check .yaml files
		if !strings.Contains(info.Name(), ".yaml") {
			return nil
		}

		// check if this file contains any image: lines that aren't using new templates
		incompatible, err := imagechecker.FileHasIncorrectImage(path, *skipComments)
		if err != nil {
			return nil
		}

		if incompatible {
			*foundIncompatible = true
		}

		return nil
	}
}
