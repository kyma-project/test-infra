package file

import (
	"io"
	"os"
	"path/filepath"
)

// FindAllRecursively finds files with defined extension recursively
func FindAllRecursively(rootPath, extension string) ([]string, error) {
	var paths []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != extension {
			return err
		}

		paths = append(paths, path)
		return nil
	})

	return paths, err
}

// ReadFile .
func ReadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	content := string(data)

	return content, nil
}
# (2025-03-04)