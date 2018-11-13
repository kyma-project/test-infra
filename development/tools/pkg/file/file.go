package file

import (
	"os"
	"path/filepath"
)

func FindAllRec(rootPath, extension string) ([]string, error) {
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
