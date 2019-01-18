package file

import (
	"io/ioutil"
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

//ReadFile .
func ReadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	content := string(data)

	return content, nil
}

// SaveDataToTmpFile writes a slice of bytes to a tmp file
func SaveDataToTmpFile(bytes []byte, tmpFilePattern string) (*os.File, error) {

	artifactFile, err := ioutil.TempFile("", tmpFilePattern)
	if err != nil {
		return nil, err
	}

	defer artifactFile.Close()

	_, err = artifactFile.Write(bytes)
	if err != nil {
		return nil, err
	}

	return artifactFile, nil
}
