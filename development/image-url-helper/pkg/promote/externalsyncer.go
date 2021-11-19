package promote

import (
	"bytes"
	"fmt"

	imagesyncer "github.com/kyma-project/test-infra/development/image-syncer/pkg"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"gopkg.in/yaml.v3"
)

func PrintExternalSyncerYaml(images []list.Image, targetContainerRegistry, targetTag string, sign bool) error {
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

func convertImageslist(images []list.Image, targetContainerRegistry, targetTag string, sign bool) imagesyncer.SyncDef {
	syncDef := imagesyncer.SyncDef{}
	syncDef.TargetRepoPrefix = targetContainerRegistry
	syncDef.Sign = sign
	for _, image := range images {
		tmpImage := imagesyncer.Image{}
		tmpImage.Source = image.FullImageURL()
		if targetTag != "" {
			tmpImage.Tag = targetTag
		}
		syncDef.Images = append(syncDef.Images, tmpImage)
	}

	return syncDef
}
