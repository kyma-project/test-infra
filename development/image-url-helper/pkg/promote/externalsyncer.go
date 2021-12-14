package promote

import (
	"bytes"
	"fmt"
	"sort"

	imagesyncer "github.com/kyma-project/test-infra/development/image-syncer/pkg"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/common"
	"gopkg.in/yaml.v3"
)

// PrintExternalSyncerYaml prints out a YAML file ready to be used by the image-syncer tool to copy images to new container registry, with option to retag them
func PrintExternalSyncerYaml(images common.ComponentImageMap, targetContainerRegistry, targetTag string, sign bool) error {
	imagesConverted := convertImageslist(images, targetContainerRegistry, targetTag, sign)

	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	err := encoder.Encode(imagesConverted)
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(out.String())
	return nil
}

// convertImageslist takes in a list of images, target repository & tag and creates a SyncDef structure that can be later marshalled and used by the image-syncer tool
func convertImageslist(images common.ComponentImageMap, targetContainerRegistry, targetTag string, sign bool) imagesyncer.SyncDef {

	imageNames := make([]string, 0)
	for _, image := range images {
		imageNames = append(imageNames, image.Image.FullImageURL())
	}
	sort.Strings(imageNames)

	syncDef := imagesyncer.SyncDef{}
	syncDef.TargetRepoPrefix = targetContainerRegistry + "/"
	syncDef.Sign = sign
	for _, fullImageURL := range imageNames {
		tmpImage := imagesyncer.Image{}
		tmpImage.Source = fullImageURL
		if targetTag != "" {
			tmpImage.Tag = targetTag
		}
		syncDef.Images = append(syncDef.Images, tmpImage)
	}

	return syncDef
}
