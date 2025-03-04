package extractimageurls

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

// ExtractFunc is standard type for function used to return image urls from reader
type ExtractFunc func(reader io.Reader) ([]string, error)

// FromFiles uses given extract function to return image urls from given files
func FromFiles(files []string, extract ExtractFunc) ([]string, error) {
	var images []string
	for _, file := range files {
		reader, err := os.Open(file)
		if err != nil {
			return nil, err
		}

		img, err := extract(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to extract images from file %s: %s", file, err)
		}

		images = append(images, img...)
	}

	return images, nil
}

// FindFilesInDirectory returns list of files that match regexp under specified directory
func FindFilesInDirectory(rootPath, regex string) ([]string, error) {
	var files []string
	//nolint:revive
	err := filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		re := regexp.MustCompile(regex)

		if re.MatchString(path) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// UniqueImages returns list of unique image urls from given list of image urls
func UniqueImages(images []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range images {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

// SplitYamlIntoSections split yaml into separated section based --- according to yaml specification
func SplitYamlIntoSections(data []byte) [][]byte {
	re := regexp.MustCompile("(?m)^---\n")

	strings := re.Split(string(data), -1)

	var result [][]byte
	for _, str := range strings {
		result = append(result, []byte(str))
	}

	return result
}
# (2025-03-04)