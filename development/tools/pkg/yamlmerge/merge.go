package yamlmerge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
)

// MergeFiles merges all files found on path with the given extension into one file, that is the target file.
func MergeFiles(path, extension, target string, changeFile bool) {
	var dryRunPrefix string
	if !changeFile {
		dryRunPrefix = "[DRYRUN] "
	}
	files := collectFiles(path, extension)

	files = removeFromArray(files, target)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("Couldn't read file (%s) contents: %s\n", f, err.Error())
		}
		if !strings.HasSuffix(string(data), "\n") { // append newline if file doesn't end with newline
			data = append(data, "\n"...)
		}

		if changeFile {
			t, fileErr := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if fileErr != nil {
				log.Fatalf("Couldn't open or create file: %s\n", fileErr.Error())
			}

			if _, writeErr := t.Write(data); writeErr != nil {
				log.Fatalf("Error writing file data: %s\n", writeErr.Error())
			}
		}
		common.Shout("%sAppending content of %s successful.", dryRunPrefix, f)
	}
}

func collectFiles(path, extension string) []string {
	var files []string

	dirInfo, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("Couldn't read files in path: %s", err.Error())
	}

	for _, f := range dirInfo {
		fileInfo, err := f.Info()
		if err != nil {
			log.Fatalf("Couldn't read file info: %s", err.Error())
		}

		if fileInfo.Mode().IsRegular() && filepath.Ext(fileInfo.Name()) == extension {
			files = append(files, fmt.Sprintf("%s%s%s", path, string(os.PathSeparator), f.Name()))
		}
	}
	return files
}

func removeFromArray(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
