package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	skippedInfo = ""

	kymaResourcesDirectory = flag.String("kymaDirectory", "/home/prow/go/src/github.com/kyma-project/kyma/resources/", "Path to Kyma resources")
	skipComments           = flag.Bool("skipComments", true, "Skip commented out lines")
	foundIncompatible      = false

	imageRegexpString  = "image: "
	imageRegexp        = regexp.MustCompile(imageRegexpString)
	commentedOutRegexp = regexp.MustCompile("#(.*)" + imageRegexpString)

	// Somehow this string breaks syntax colouring in Visual Studio Code, this is why I do this addition to still see what I do
	includeRegexpString = "{{ " + "include \"(short)?imageurl\"(.*)"
	newWayRegexp        = regexp.MustCompile(includeRegexpString)
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
	incompatible, err := fileHasIncorrectImage(path)
	if err != nil {
		return nil
	}

	if incompatible {
		foundIncompatible = true
	}

	return nil
}

func fileHasIncorrectImage(path string) (bool, error) {
	incompatible := false
	//open file and read it line by line
	fileHandle, err := os.Open(path)
	if err != nil {
		return true, err
	}
	defer fileHandle.Close()

	lineNumber := 1
	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		if oldImageFormat(scanner.Text()) {
			incompatible = true
			fmt.Printf("%s:%d: %s\n", path, lineNumber, strings.Trim(scanner.Text(), " "))
		}
		lineNumber += 1
	}

	if err := scanner.Err(); err != nil {
		return true, err
	}

	return incompatible, nil
}

func oldImageFormat(line string) bool {
	// skip all uninteresting lines and just "name:" in its own line
	if imageRegexp.MatchString(line) {
		// check if we should ship commented out lines or not
		if !*skipComments || !commentedOutRegexp.MatchString(line) {
			// and if there is our template
			if !newWayRegexp.MatchString(line) {
				return true
			}
		}
	}
	return false
}
