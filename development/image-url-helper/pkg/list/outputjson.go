package list

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/common"
	"gopkg.in/yaml.v2"
)

// CustomFields contains list of image custom fields
type CustomFields struct {
	Components string `json:"components" yaml:"components"`
	Image      string `json:"image" yaml:"image"`
}

// OutputImage describes an image in the output format
type OutputImage struct {
	Name         string       `json:"name" yaml:"name"`
	CustomFields CustomFields `json:"custom_fields" yaml:"custom_fields"`
}

// OutputImageList contains a list of images in the output format
type OutputImageList struct {
	Images []OutputImage `json:"images" yaml:"images"`
}

// PrintImagesJSON prints JSON list with names and components for each image
func PrintImagesJSON(allImages common.ImageMap, imageComponentsMap common.ImageToComponents) error {
	imagesConverted := convertImageslist(allImages, imageComponentsMap)

	out, err := json.MarshalIndent(imagesConverted, "", "  ")
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(string(out))
	return nil
}

// PrintImagesYAML prints YAML list with names and components for each image
func PrintImagesYAML(allImages common.ImageMap, imageComponentsMap common.ImageToComponents) error {
	imagesConverted := convertImageslist(allImages, imageComponentsMap)

	out, err := yaml.Marshal(imagesConverted)
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(string(out))
	return nil
}

// convertImageslist takes in a list of images and image to component mapping and creates an OutputImageList structure that can be later marshalled and used by the security scan tool
func convertImageslist(allImages common.ImageMap, imageComponentsMap common.ImageToComponents) OutputImageList {
	imagesConverted := OutputImageList{}

	imageNames := make([]string, 0)
	for _, image := range allImages {
		imageNames = append(imageNames, image.FullImageURL())
	}
	sort.Strings(imageNames)

	for _, fullImageURL := range imageNames {
		imageTmp := OutputImage{}
		imageTmp.Name = fullImageURL
		imageTmp.CustomFields.Image = fullImageURL
		components := imageComponentsMap[fullImageURL]
		imageTmp.CustomFields.Components = strings.Join(components, ",")
		imagesConverted.Images = append(imagesConverted.Images, imageTmp)
	}

	return imagesConverted
}
