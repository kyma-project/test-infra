package jobsuite

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// Suite interface implements Run function
type Suite interface {
	Run(t *testing.T)
}

// JobConfigPathProvider interface provides function to provide JobConfigPath
type JobConfigPathProvider interface {
	JobConfigPath() string
}

// CheckFilesAreTested function
func CheckFilesAreTested(repos map[string]struct{}, testedConfigurations map[string]struct{}, jobBasePath string, subfolders []string) func(t *testing.T) {
	return func(t *testing.T) {
		for repo := range repos {
			var err error
			found := false
			for _, subfolder := range subfolders {
				folderToCheck := path.Join(jobBasePath, path.Base(repo), subfolder)
				err := filepath.Walk(folderToCheck,
					func(fp string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if info.IsDir() {
							return nil
						}
						if path.Ext(fp) != ".yaml" {
							return nil
						}

						if !strings.Contains(path.Base(fp), "generic") {
							return nil
						}

						if _, tested := testedConfigurations[fp]; !tested {
							t.Errorf("File %v was not covered by test", fp)
						}
						return nil

					})
				if err == nil {
					found = true
				}
			}
			if !found {
				t.Errorf("%v", err)
			}
		}
	}
}
