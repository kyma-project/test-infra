package pkg

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
)

func ExtractImagesFromFiles(files []string, extract func(path string) ([]string, error)) ([]string, error) {
	var images []string
	for _, file := range files {
		img, err := extract(file)
		if err != nil {
			return nil, fmt.Errorf("failed to extract images from file %s: %s", file, err)
		}

		images = append(images, img...)
	}

	return images, nil
}

func FindFilesInDirectory(rootPath, regex string) ([]string, error) {
	var files []string
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
