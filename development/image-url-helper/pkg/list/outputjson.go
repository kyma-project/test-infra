package list

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

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
func PrintImagesJSON(allImages ImageList, imageComponentsMap ImageToComponents) error {
	sort.Sort(allImages)
	imagesConverted := convertimageslist(allImages, imageComponentsMap)

	out, err := json.MarshalIndent(imagesConverted, "", "  ")
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(string(out))
	return nil
}

// PrintImagesYAML prints YAML list with names and components for each image
func PrintImagesYAML(allImages ImageList, imageComponentsMap ImageToComponents) error {
	sort.Sort(allImages)
	imagesConverted := convertimageslist(allImages, imageComponentsMap)

	out, err := yaml.Marshal(imagesConverted)
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err)
	}
	fmt.Println(string(out))
	return nil
}

// convertImageslist takes in a list of images and image to component mapping and creates an OutputImageList structure that can be later marshalled and used by the security scan tool
func convertimageslist(allImages ImageList, imageComponentsMap ImageToComponents) OutputImageList {
	imagesConverted := OutputImageList{}

	for _, image := range allImages {
		imageTmp := OutputImage{}
		imageTmp.Name = image.FullImageURL()
		imageTmp.CustomFields.Image = image.FullImageURL()
		components := imageComponentsMap[image.FullImageURL()]
		imageTmp.CustomFields.Components = strings.Join(components, ",")
		imagesConverted.Images = append(imagesConverted.Images, imageTmp)
	}

	return imagesConverted
}
