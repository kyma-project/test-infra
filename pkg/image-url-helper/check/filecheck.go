package check

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
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

// Exclude contains excluded image values for a given file
type Exclude struct {
	Filename string   `yaml:"filename"`
	Images   []string `yaml:"images"`
}

// excludesList contains a list of excluded image values per each file
type excludesList struct {
	Excludes []Exclude `yaml:"excludes"`
}

// ParseExcludes reads the exclude list file and returns list of excludes
func ParseExcludes(excludesListFilename string) ([]Exclude, error) {
	if excludesListFilename == "" {
		return nil, nil
	}

	excludesListFile, err := os.ReadFile(excludesListFilename)
	if err != nil {
		return nil, err
	}

	var excludesList excludesList

	if err = yaml.Unmarshal(excludesListFile, &excludesList); err != nil {
		return nil, err
	}

	return excludesList.Excludes, nil
}

func GetkWalkFunc(resourcesDirectory string, imagesDefinedOutside *[]ImageLine, skipComments bool, excludesList []Exclude) filepath.WalkFunc {
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
		incompatibleLines, err := FileHasIncorrectImage(resourcesDirectory, path, skipComments, excludesList)
		if err != nil {
			return nil
		}
		*imagesDefinedOutside = append(*imagesDefinedOutside, incompatibleLines...)

		return nil
	}
}

// FileHasIncorrectImage checks if the file contains lines that doesn't use new template for images
func FileHasIncorrectImage(resourcesDirectory, path string, skipComments bool, excludesList []Exclude) ([]ImageLine, error) {
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
		if newImageFormat(scanner.Text(), skipComments) {
			if !imageInExcludeList(resourcesDirectory, path, scanner.Text(), excludesList) {
				incompatible = append(incompatible, ImageLine{Filename: path, LineNumber: lineNumber, Line: strings.Trim(scanner.Text(), " ")})
			}
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return incompatible, nil
}

// imageInExcludeList checks if the image value in the given line is on the excludes list
func imageInExcludeList(resourcesDirectory, filename, line string, excludesList []Exclude) bool {
	for _, exclude := range excludesList {
		if strings.Replace(filename, resourcesDirectory+"/", "", -1) == exclude.Filename {
			for _, image := range exclude.Images {
				// naive line parsing
				parsedImage := strings.Replace(line, "image: ", "", -1)
				parsedImage = strings.Replace(parsedImage, "\"", "", -1)
				parsedImage = strings.Trim(parsedImage, " ")
				if strings.HasPrefix(parsedImage, image) {
					return true
				}
			}
		}
	}
	return false
}

// newImageFormat checks and prints lines that doesn't use the "imageurl" or "shortimageurl" template
func newImageFormat(line string, skipComments bool) bool {
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
# (2025-03-04)