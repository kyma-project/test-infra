package list

import (
	"github.com/kyma-project/test-infra/pkg/image-url-helper/common"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func GetWalkFunc(resourcesDirectory string, images, testImages common.ComponentImageMap) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		//pass the error further, this shouldn't ever happen
		if err != nil {
			return err
		}

		// skip directory entries, we just want files
		if info.IsDir() {
			return nil
		}

		// we only want to check values.yaml files
		if info.Name() != "values.yaml" {
			return nil
		}

		var parsedFile common.ValueFile

		yamlFile, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(yamlFile, &parsedFile)
		if err != nil {
			return err
		}

		component := strings.Replace(path, resourcesDirectory+"/", "", -1)
		component = strings.Replace(component, "/values.yaml", "", -1)

		common.AppendImagesToMap(parsedFile, images, testImages, component)

		return nil
	}
}
# (2025-03-04)