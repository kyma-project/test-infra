package promote

import (
	"bytes"
	"fmt"

	imagesyncer "github.com/kyma-project/test-infra/development/image-syncer/pkg"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"gopkg.in/yaml.v3"
)

func appendImagesToList(parsedFile list.ValueFile, images *[]list.Image) {
	for _, image := range parsedFile.Global.Images {
		// add registry info directly into the image struct
		if image.ContainerRegistryPath == "" {
			image.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
		}
		// remove duplicates
		if !list.ImageListContains(*images, image) {
			*images = append(*images, image)
		}
	}

	for _, testImage := range parsedFile.Global.TestImages {
		if testImage.ContainerRegistryPath == "" {
			testImage.ContainerRegistryPath = parsedFile.Global.ContainerRegistry.Path
		}
		if !list.ImageListContains(*images, testImage) {
			*images = append(*images, testImage)
		}
	}
}

func PrintExternalSyncerYaml(images []list.Image, targetContainerRegistry, targetTag string, sign bool) error {
	imagesConverted := convertimageslist(images, targetContainerRegistry, targetTag, sign)

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

func convertimageslist(images []list.Image, targetContainerRegistry, targetTag string, sign bool) imagesyncer.SyncDef {
	syncDef := imagesyncer.SyncDef{}
	syncDef.TargetRepoPrefix = targetContainerRegistry
	syncDef.Sign = sign
	for _, image := range images {
		tmpImage := imagesyncer.Image{}
		tmpImage.Source = image.String()
		if targetTag != "" {
			tmpImage.Tag = targetTag
		}
		syncDef.Images = append(syncDef.Images, tmpImage)
	}

	return syncDef
}
