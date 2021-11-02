package check

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
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

// ImageLine defines a line inside a file that doesn't use new template for image deifnition
type ImageLine struct {
	Filename   string
	LineNumber int
	Line       string
}

func GetkWalkFunc(imagesDefinedOutside *[]ImageLine, skipComments bool) filepath.WalkFunc {
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
		incompatibleLines, err := FileHasIncorrectImage(path, skipComments)
		if err != nil {
			return nil
		}
		*imagesDefinedOutside = append(*imagesDefinedOutside, incompatibleLines...)

		return nil
	}
}

// FileHasIncorrectImage checks if the file contains lines that doesn't use new template for images
func FileHasIncorrectImage(path string, skipComments bool) ([]ImageLine, error) {
	var incompatible []ImageLine
	//open file and read it line by line
	fileHandle, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	lineNumber := 1
	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		if oldImageFormat(scanner.Text(), skipComments) {
			incompatible = append(incompatible, ImageLine{Filename: path, LineNumber: lineNumber, Line: strings.Trim(scanner.Text(), " ")})
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
