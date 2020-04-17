package jobsuite

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

type Suite interface {
	Run(t *testing.T)
}

type JobConfigPathProvider interface {
	JobConfigPath() string
}

func CheckFilesAreTested(repos map[string]struct{}, testedConfigurations map[string]struct{}, jobBasePath, subfolder string) func(t *testing.T) {
	return func(t *testing.T) {
		for repo, _ := range repos {
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
			if err != nil {
				t.Errorf("%v", err)
			}
		}
	}
}
