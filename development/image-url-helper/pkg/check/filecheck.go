package check

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	imageRegexpString  = "image: "
	imageRegexp        = regexp.MustCompile(imageRegexpString)
	commentedOutRegexp = regexp.MustCompile("#(.*)" + imageRegexpString)

	// "{{" breaks syntax colouring in Visual Studio Code, The comment at the end prevents that
	includeRegexpString = "{{\\s?include \"(short)?imageurl\"(.*)" // }}"
	newWayRegexp        = regexp.MustCompile(includeRegexpString)
)

// FileHasIncorrectImage checks if the file contains lines that doesn't use new template for images
func FileHasIncorrectImage(path string, skipComments bool) (bool, error) {
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
		if oldImageFormat(scanner.Text(), skipComments) {
			incompatible = true
			fmt.Printf("%s:%d: %s\n", path, lineNumber, strings.Trim(scanner.Text(), " "))
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return true, err
	}

	return incompatible, nil
}

// oldImageFormat checks and prints lines that uses old image: format
func oldImageFormat(line string, skipComments bool) bool {
	// skip all uninteresting lines and just "name:" in its own line
	if imageRegexp.MatchString(line) {
		// check if we should ship commented out lines or not
		if !skipComments || !commentedOutRegexp.MatchString(line) {
			// and if there is our template
			if !newWayRegexp.MatchString(line) {
				return true
			}
		}
	}
	return false
}
